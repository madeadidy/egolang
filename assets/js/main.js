$(function () {
  // ================== SLIDER RANGE ====================
  $("#slider-range").slider({
    range: true,
    min: 0,
    max: 2500,
    values: [10, 2500],
    slide: function (event, ui) {
      $("#amount").val("$" + ui.values[0] + " - $" + ui.values[1]);
    },
  });

  $("#amount").val("$" + $("#slider-range").slider("values", 0) + " - $" + $("#slider-range").slider("values", 1));

  // ================== FORMAT RUPIAH ====================
  function formatRupiah(angka) {
    if (!angka || isNaN(angka)) return "Rp 0";
    return "Rp " + angka.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ".");
  }

  // ================== FORMAT ANGKA DI PAGE LOAD ====================
  function formatAllPricesOnLoad() {
    $(".price").each(function () {
      let angka = $(this).text().replace(/\./g, "").replace("Rp ", "") || "0";
      $(this).text(formatRupiah(parseInt(angka)));
    });

    let subtotal = $("#cart-subtotal").text().replace(/\./g, "").replace("Rp ", "") || "0";
    let tax = $("#cart-tax").text().replace(/\./g, "").replace("Rp ", "") || "0";
    let grandTotal = $("#grand-total").text().replace(/\./g, "").replace("Rp ", "") || "0";

    $("#cart-subtotal").text(formatRupiah(parseInt(subtotal)));
    $("#cart-tax").text(formatRupiah(parseInt(tax)));
    $("#grand-total").text(formatRupiah(parseInt(grandTotal)));
  }

  // Panggil saat halaman load
  formatAllPricesOnLoad();

  // ================== PROVINSI -> KOTA ====================
  $(".province_id").change(function () {
    let provinceID = $(this).val();
    $(".city_id").prop("disabled", true);

    $.ajax({
      url: "/carts/cities?province_id=" + provinceID,
      method: "GET",
      success: function (result) {
        $(".city_id").empty().append('<option value="">Pilih Kota / Kabupaten</option>');
        $.each(result.data, function (i, city) {
          $(".city_id").append(`<option value="${city.id}">${city.name}</option>`);
        });
        $(".city_id").prop("disabled", false);
      },
      error: function () {
        console.error("Gagal mengambil data kota");
      },
    });
  });

  // ================== KOTA -> ONGKIR ====================
  $(".city_id").change(function () {
    let cityID = $(this).val();
    let courier = $(".courier").val();

    $.ajax({
      url: "/carts/calculate-shipping",
      method: "POST",
      contentType: "application/json",
      data: JSON.stringify({
        city_id: cityID,
        courier: courier,
      }),
      success: function (result) {
        $(".shipping_fee_options").empty().append('<option value="">Pilih Paket</option>');

        if (result.data && result.data.length > 0) {
          $.each(result.data, function (i, shipping_fee_option) {
            $(".shipping_fee_options").append(
              `<option value="${shipping_fee_option.service}-${shipping_fee_option.fee}">
                ${shipping_fee_option.service} - ${formatRupiah(shipping_fee_option.fee)}
              </option>`,
            );
          });
        } else {
          $(".shipping_fee_options").append('<option value="">Tidak ada paket tersedia</option>');
        }
      },
      error: function (xhr) {
        console.error("AJAX Error:", xhr.responseText);
      },
    });
  });

  // ================== APPLY ONGKIR & UPDATE TOTAL ====================
  $(".shipping_fee_options").change(function () {
    let cityID = $(".city_id").val();
    let courier = $(".courier").val();
    let shippingFeeData = $(this).val();
    if (!shippingFeeData) return;

    let shippingPackage = shippingFeeData.split("-")[0];
    let shippingFee = parseInt(shippingFeeData.split("-")[1]);

    $.ajax({
      url: "/carts/apply-shipping",
      method: "POST",
      contentType: "application/json",
      data: JSON.stringify({
        shipping_package: shippingPackage,
        city_id: cityID,
        courier: courier,
      }),
      success: function (result) {
        // Ambil angka dari elemen (hapus titik & Rp)
        let rawSubtotal = $("#cart-subtotal").text().replace(/\./g, "").replace("Rp ", "") || "0";
        let rawTax = $("#cart-tax").text().replace(/\./g, "").replace("Rp ", "") || "0";

        let subtotal = parseInt(rawSubtotal);
        let tax = parseInt(rawTax);
        let grandTotal = subtotal + tax + shippingFee;

        // Update tampilan dengan format rupiah
        $("#cart-subtotal").text(formatRupiah(subtotal));
        $("#cart-tax").text(formatRupiah(tax));
        $("#grand-total").text(formatRupiah(grandTotal));
      },
      error: function () {
        $("#shipping-calculation-msg").html(`<div class="alert alert-warning">Pemilihan paket ongkir gagal!</div>`);
      },
    });
  });

  $(function () {
    // Event tombol +
    $(".btn-qty-plus").click(function () {
      const targetID = $(this).data("target");
      const input = $("#qty-" + targetID);
      const current = parseInt(input.val());
      input.val(current + 1);
    });

    // Event tombol -
    $(".btn-qty-minus").click(function () {
      const targetID = $(this).data("target");
      const input = $("#qty-" + targetID);
      const current = parseInt(input.val());
      if (current > 1) {
        input.val(current - 1);
      }
    });
  });

  // Delegated handlers for profile modal selects (addrProvince / addrCity)
  $(document).on("change", "#addrProvince", function () {
    var provinceID = $(this).val();
    var $city = $("#addrCity");
    $city.prop("disabled", true).empty().append('<option value="">Memuat...</option>');
    if (!provinceID) {
      $city.empty().append('<option value="">--</option>');
      return;
    }
    $.ajax({
      url: "/carts/cities?province_id=" + encodeURIComponent(provinceID),
      method: "GET",
      success: function (result) {
        $city.empty().append('<option value="">Pilih Kota</option>');
        if (result && result.data) {
          $.each(result.data, function (i, city) {
            $city.append('<option value="' + (city.id || city.ID || "") + '">' + (city.name || city.Name || "") + "</option>");
          });
        }
        $city.prop("disabled", false);
      },
      error: function (xhr) {
        console.error("Gagal mengambil data kota", xhr && xhr.responseText);
        $city.empty().append('<option value="">Gagal memuat kota</option>');
      },
    });
  });

  $(document).on("change", "#addrCity", function () {
    var cityID = $(this).val();
    var $district = $("#addrDistrict");
    var $postcode = $("#addrPostcode");
    $district.empty().append('<option value="">Memuat...</option>');
    $postcode.empty().append('<option value="">Memuat...</option>');
    if (!cityID) {
      $district.empty().append('<option value="">--</option>');
      $postcode.empty().append('<option value="">--</option>');
      return;
    }
    $.getJSON("/carts/districts?city_id=" + encodeURIComponent(cityID))
      .done(function (res) {
        $district.empty().append('<option value="">Pilih Kecamatan</option>');
        if (res && res.data) {
          $.each(res.data, function (i, d) {
            $district.append('<option value="' + (d.id || d.ID || "") + '">' + (d.name || d.Name || "") + "</option>");
          });
        }
      })
      .fail(function (xhr) {
        console.error("Gagal memuat kecamatan", xhr && xhr.responseText);
        $district.empty().append('<option value="">Gagal memuat kecamatan</option>');
      });

    $.getJSON("/carts/postcodes?city_id=" + encodeURIComponent(cityID))
      .done(function (res) {
        $postcode.empty().append('<option value="">Pilih Kode Pos</option>');
        var $manual = $("#addrPostcodeInput");
        if (res && res.data && res.data.length > 0) {
          // populate select and ensure manual input is hidden
          $.each(res.data, function (i, p) {
            var code = p.code || p.Code || "";
            $postcode.append('<option value="' + code + '">' + code + "</option>");
          });
          $postcode.show().prop("disabled", false);
          if ($manual && $manual.length) {
            $manual.hide().prop("disabled", true);
          }
        } else {
          // no postcodes -> show manual input and disable select
          $postcode.hide().prop("disabled", true);
          if ($manual && $manual.length) {
            $manual.show().prop("disabled", false);
          }
        }
      })
      .fail(function (xhr) {
        console.error("Gagal memuat kode pos", xhr && xhr.responseText);
        var $manual = $("#addrPostcodeInput");
        $postcode.empty().append('<option value="">Gagal memuat kode pos</option>').hide().prop("disabled", true);
        if ($manual && $manual.length) {
          $manual.show().prop("disabled", false);
        }
      });
  });
});

<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Products</h1>

    <section class="mb-4">
      <h2 class="font-semibold">Buat Produk Baru</h2>
      <form @submit.prevent="createProduct" class="mt-2">
        <input v-model="form.name" placeholder="Nama" required class="border px-2 py-1 mr-2" />
        <input v-model.number="form.price" placeholder="Harga" required class="border px-2 py-1 mr-2" />
        <input v-model.number="form.stock" placeholder="Stok" required class="border px-2 py-1 mr-2" />
        <button type="submit" :disabled="savingCreate" class="px-3 py-1 bg-blue-600 text-white">
          <span v-if="savingCreate">Menyimpan...</span>
          <span v-else>Buat</span>
        </button>
      </form>
    </section>

    <div v-if="statusMessage" class="mb-2 text-green-700">{{ statusMessage }}</div>
    <div v-if="errorMessage" class="mb-2 text-red-700">{{ errorMessage }}</div>

    <div v-if="loading">Memuat produk...</div>
    <div v-else>
      <table class="min-w-full border-collapse">
        <thead>
          <tr>
            <th class="border px-2 py-1">ID</th>
            <th class="border px-2 py-1">Nama</th>
            <th class="border px-2 py-1">Harga</th>
            <th class="border px-2 py-1">Stok</th>
            <th class="border px-2 py-1">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in products" :key="p.ID || p.id">
            <td class="border px-2 py-1">{{ p.ID || p.id }}</td>
            <td class="border px-2 py-1">{{ p.Name || p.name }}</td>
            <td class="border px-2 py-1">{{ displayPrice(p.Price) }}</td>
            <td class="border px-2 py-1">{{ p.Stock || p.stock }}</td>
            <td class="border px-2 py-1">
              <button @click="startEdit(p)" class="mr-2 px-2 py-1 bg-yellow-500 text-white">Edit</button>
              <button @click="remove(p)" class="px-2 py-1 bg-red-600 text-white">Hapus</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="editing" class="mt-4 p-3 border">
      <h3 class="font-semibold">Edit Produk</h3>
      <form @submit.prevent="updateProduct" class="mt-2">
        <input v-model="form.name" placeholder="Nama" required class="border px-2 py-1 mr-2" />
        <input v-model.number="form.price" placeholder="Harga" required class="border px-2 py-1 mr-2" />
        <input v-model.number="form.stock" placeholder="Stok" required class="border px-2 py-1 mr-2" />
        <button type="submit" :disabled="savingUpdate" class="px-3 py-1 bg-green-600 text-white mr-2">
          <span v-if="savingUpdate">Menyimpan...</span>
          <span v-else>Simpan</span>
        </button>
        <button type="button" @click="cancelEdit" class="px-3 py-1 bg-gray-400">Batal</button>
      </form>
      <div class="mt-3">
        <label class="block mb-1 font-medium">Upload Gambar Produk</label>
        <input type="file" @change="onFileChange" accept="image/*" />
        <button @click="uploadImage" :disabled="uploading || !currentId" class="mt-2 px-3 py-1 bg-blue-600 text-white">
          <span v-if="uploading">Mengunggah...</span>
          <span v-else>Unggah</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from "vue";

export default defineComponent({
  name: "ProductsView",
  setup() {
    const products = ref<any[]>([]);
    const loading = ref(true);
    const editing = ref(false);
    const currentId = ref<string | null>(null);
    const form = ref({ name: "", price: 0, stock: 0 });
    const savingCreate = ref(false);
    const savingUpdate = ref(false);
    const deletingId = ref<string | null>(null);
    const selectedFile = ref<File | null>(null);
    const uploading = ref(false);
    const statusMessage = ref("");
    const errorMessage = ref("");

    function authHeaders() {
      const headers: any = {};
      if ((window as any).ADMIN_API_KEY) headers["Authorization"] = "Bearer " + (window as any).ADMIN_API_KEY;
      return headers;
    }

    async function fetchProducts() {
      loading.value = true;
      try {
        const res = await fetch("/api/admin/products", { headers: authHeaders() });
        if (!res.ok) throw new Error(res.statusText || String(res.status));
        products.value = await res.json();
      } catch (err: any) {
        console.error("fetch products error", err);
        errorMessage.value = err.message || "Gagal memuat produk";
      } finally {
        loading.value = false;
      }
    }

    async function createProduct() {
      statusMessage.value = "";
      errorMessage.value = "";
      if (!form.value.name || form.value.name.length < 2) {
        errorMessage.value = "Nama minimal 2 karakter";
        return;
      }
      if (Number(form.value.price) <= 0) {
        errorMessage.value = "Harga harus lebih dari 0";
        return;
      }
      savingCreate.value = true;
      try {
        const res = await fetch("/api/admin/products", {
          method: "POST",
          headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
          body: JSON.stringify({ name: form.value.name, price: form.value.price, stock: form.value.stock }),
        });
        if (!res.ok) throw new Error("create failed");
        const p = await res.json();
        products.value.unshift(p);
        form.value = { name: "", price: 0, stock: 0 };
        statusMessage.value = "Produk berhasil dibuat";
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal membuat produk";
      } finally {
        savingCreate.value = false;
      }
    }

    function startEdit(p: any) {
      editing.value = true;
      currentId.value = p.ID || p.id;
      form.value = { name: p.Name || p.name || "", price: parseFloat((p.Price && (p.Price.String || p.Price)) || 0), stock: p.Stock || p.stock || 0 };
    }

    function cancelEdit() {
      editing.value = false;
      currentId.value = null;
      form.value = { name: "", price: 0, stock: 0 };
    }

    async function updateProduct() {
      if (!currentId.value) return;
      statusMessage.value = "";
      errorMessage.value = "";
      if (!form.value.name || form.value.name.length < 2) {
        errorMessage.value = "Nama minimal 2 karakter";
        return;
      }
      if (Number(form.value.price) <= 0) {
        errorMessage.value = "Harga harus lebih dari 0";
        return;
      }
      savingUpdate.value = true;
      try {
        const res = await fetch("/api/admin/products/" + currentId.value, {
          method: "PUT",
          headers: Object.assign({ "Content-Type": "application/json" }, authHeaders()),
          body: JSON.stringify({ name: form.value.name, price: form.value.price, stock: form.value.stock }),
        });
        if (!res.ok) throw new Error("update failed");
        const updated = await res.json();
        const idx = products.value.findIndex((x) => (x.ID || x.id) === currentId.value);
        if (idx !== -1) products.value.splice(idx, 1, updated);
        cancelEdit();
        statusMessage.value = "Produk diperbarui";
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal memperbarui produk";
      } finally {
        savingUpdate.value = false;
      }
    }

    function onFileChange(e: Event) {
      const input = e.target as HTMLInputElement;
      if (input.files && input.files.length > 0) {
        selectedFile.value = input.files[0];
      } else {
        selectedFile.value = null;
      }
    }

    async function uploadImage() {
      if (!currentId.value) return;
      if (!selectedFile.value) {
        errorMessage.value = "Pilih file gambar terlebih dahulu";
        return;
      }
      uploading.value = true;
      errorMessage.value = "";
      statusMessage.value = "";
      try {
        const fd = new FormData();
        fd.append("image", selectedFile.value as Blob);
        const headers: any = {};
        if ((window as any).ADMIN_API_KEY) headers["Authorization"] = "Bearer " + (window as any).ADMIN_API_KEY;
        const res = await fetch("/api/admin/products/" + currentId.value + "/images", { method: "POST", headers, body: fd });
        if (!res.ok) throw new Error("upload failed");
        // refresh product
        const r2 = await fetch("/api/admin/products/" + currentId.value);
        if (r2.ok) {
          const updated = await r2.json();
          const idx = products.value.findIndex((x) => (x.ID || x.id) === currentId.value);
          if (idx !== -1) products.value.splice(idx, 1, updated);
        }
        statusMessage.value = "Gambar berhasil diunggah";
        selectedFile.value = null;
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal mengunggah gambar";
      } finally {
        uploading.value = false;
      }
    }

    async function remove(p: any) {
      const id = p.ID || p.id;
      if (!confirm("Hapus produk ini?")) return;
      deletingId.value = id;
      errorMessage.value = "";
      statusMessage.value = "";
      try {
        const res = await fetch("/api/admin/products/" + id, { method: "DELETE", headers: authHeaders() });
        if (res.status !== 204) throw new Error("delete failed");
        const idx = products.value.findIndex((x) => (x.ID || x.id) === id);
        if (idx !== -1) products.value.splice(idx, 1);
        statusMessage.value = "Produk dihapus";
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal menghapus produk";
      } finally {
        deletingId.value = null;
      }
    }

    function displayPrice(p: any) {
      if (!p) return "";
      if (typeof p === "object" && p !== null && "String" in p) return p.String || "";
      if (typeof p === "string") return p;
      return String(p);
    }

    onMounted(() => {
      fetchProducts();
    });

    return {
      products,
      loading,
      form,
      createProduct,
      displayPrice,
      editing,
      startEdit,
      cancelEdit,
      updateProduct,
      remove,
      savingCreate,
      savingUpdate,
      deletingId,
      statusMessage,
      errorMessage,
      currentId,
      selectedFile,
      uploading,
      onFileChange,
      uploadImage,
    };
  },
});
</script>

<style scoped>
.border {
  border: 1px solid #e5e7eb;
}
.px-2 {
  padding-left: 8px;
  padding-right: 8px;
}
.py-1 {
  padding-top: 4px;
  padding-bottom: 4px;
}
.text-2xl {
  font-size: 1.5rem;
}
.font-bold {
  font-weight: 700;
}
.mb-4 {
  margin-bottom: 1rem;
}
</style>

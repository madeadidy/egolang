<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Orders</h1>

    <div v-if="loading" class="flex items-center gap-2">
      <span class="spinner" aria-hidden="true"></span>
      <span>Memuat pesanan...</span>
    </div>

    <div v-if="errorMessage" class="mb-2 text-red-700">{{ errorMessage }}</div>

    <div v-if="orders && orders.length">
      <table class="min-w-full border-collapse">
        <thead>
          <tr>
            <th class="border px-2 py-1">ID</th>
            <th class="border px-2 py-1">Code</th>
            <th class="border px-2 py-1">Customer</th>
            <th class="border px-2 py-1">Order Date</th>
            <th class="border px-2 py-1">Items</th>
            <th class="border px-2 py-1">Total</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in orders" :key="o.id">
            <td class="border px-2 py-1">{{ o.id }}</td>
            <td class="border px-2 py-1">{{ o.code }}</td>
            <td class="border px-2 py-1">{{ (o.user && o.user.first_name + " " + o.user.last_name) || o.user_email || "-" }}</td>
            <td class="border px-2 py-1">{{ formatDate(o.order_date) }}</td>
            <td class="border px-2 py-1">{{ o.items_count || (o.order_items && o.order_items.length) || 0 }}</td>
            <td class="border px-2 py-1">{{ o.grand_total || o.total || "-" }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else-if="!loading">
      <p>Tidak ada data pesanan via API. Anda dapat membuka halaman server-side <a class="text-primary" href="/admin/orders">/admin/orders</a> untuk melihat daftar lengkap.</p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from "vue";

export default defineComponent({
  name: "OrdersView",
  setup() {
    const orders = ref<any[] | null>(null);
    const loading = ref(true);
    const errorMessage = ref("");
    const adminKey = (window as any).__ADMIN_API_KEY || (window as any).ADMIN_API_KEY || "";

    async function fetchOrders() {
      loading.value = true;
      errorMessage.value = "";
      try {
        const headers: any = {};
        if (adminKey) headers["Authorization"] = "Bearer " + adminKey;
        const res = await fetch("/api/admin/orders", { headers });
        if (!res.ok) {
          // API might not exist - set orders to empty and show hint
          orders.value = [];
          return;
        }
        orders.value = await res.json();
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal memuat pesanan";
        orders.value = [];
      } finally {
        loading.value = false;
      }
    }

    function formatDate(v: any) {
      if (!v) return "";
      try {
        return new Date(v).toLocaleString();
      } catch {
        return String(v);
      }
    }

    onMounted(() => {
      fetchOrders();
    });
    return { orders, loading, errorMessage, formatDate };
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
.spinner {
  border: 3px solid rgba(0, 0, 0, 0.08);
  border-top-color: #2563eb;
  border-radius: 50%;
  width: 18px;
  height: 18px;
  animation: spin 0.9s linear infinite;
  display: inline-block;
  vertical-align: middle;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

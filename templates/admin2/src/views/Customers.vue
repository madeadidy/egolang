<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Customers</h1>

    <div v-if="loading" class="flex items-center gap-2">
      <span class="spinner" aria-hidden="true"></span>
      <span>Memuat pelanggan...</span>
    </div>

    <div v-if="errorMessage" class="mb-2 text-red-700">{{ errorMessage }}</div>

    <div v-else>
      <table class="min-w-full border-collapse">
        <thead>
          <tr>
            <th class="border px-2 py-1">ID</th>
            <th class="border px-2 py-1">Nama</th>
            <th class="border px-2 py-1">Email</th>
            <th class="border px-2 py-1">Role</th>
            <th class="border px-2 py-1">Terdaftar</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td class="border px-2 py-1">{{ u.id }}</td>
            <td class="border px-2 py-1">{{ u.first_name }} {{ u.last_name }}</td>
            <td class="border px-2 py-1">{{ u.email }}</td>
            <td class="border px-2 py-1">{{ u.role }}</td>
            <td class="border px-2 py-1">{{ formatDate(u.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from "vue";

export default defineComponent({
  name: "CustomersView",
  setup() {
    const users = ref<any[]>([]);
    const loading = ref(true);
    const errorMessage = ref("");
    const adminKey = (window as any).__ADMIN_API_KEY || (window as any).ADMIN_API_KEY || "";

    async function fetchUsers() {
      loading.value = true;
      errorMessage.value = "";
      try {
        const headers: any = {};
        if (adminKey) headers["Authorization"] = "Bearer " + adminKey;
        const res = await fetch("/api/admin/users", { headers });
        if (!res.ok) throw new Error(res.statusText);
        users.value = await res.json();
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal memuat pelanggan";
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
      fetchUsers();
    });
    return { users, loading, errorMessage, formatDate };
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

<template>
  <div class="p-4">
    <h1 class="text-2xl font-bold mb-4">Manajemen User</h1>

    <div v-if="statusMessage" class="mb-2 text-green-700">{{ statusMessage }}</div>
    <div v-if="errorMessage" class="mb-2 text-red-700">{{ errorMessage }}</div>

    <div v-if="loading" class="flex items-center gap-2">
      <span class="spinner" aria-hidden="true"></span>
      <span>Memuat pengguna...</span>
    </div>
    <div v-else>
      <table class="min-w-full border-collapse">
        <thead>
          <tr>
            <th class="border px-2 py-1">ID</th>
            <th class="border px-2 py-1">Nama</th>
            <th class="border px-2 py-1">Email</th>
            <th class="border px-2 py-1">Role</th>
            <th class="border px-2 py-1">Terdaftar</th>
            <th class="border px-2 py-1">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td class="border px-2 py-1">{{ u.id }}</td>
            <td class="border px-2 py-1">{{ u.first_name }} {{ u.last_name }}</td>
            <td class="border px-2 py-1">{{ u.email }}</td>
            <td class="border px-2 py-1">
              <select v-model="u.role" @change="changeRole(u)" :disabled="updatingId === u.id" class="border px-2 py-1">
                <option value="user">user</option>
                <option value="admin">admin</option>
                <option value="superadmin">superadmin</option>
              </select>
              <span v-if="updatingId === u.id" class="ml-2 spinner" aria-hidden="true"></span>
            </td>
            <td class="border px-2 py-1">{{ formatDate(u.created_at) }}</td>
            <td class="border px-2 py-1">
              <button @click="confirmDelete(u)" :disabled="deletingId === u.id" class="px-2 py-1 bg-red-600 text-white">
                <span v-if="deletingId === u.id" class="spinner" aria-hidden="true"></span>
                <span v-else>Hapus</span>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from "vue";

export default defineComponent({
  name: "UsersView",
  setup() {
    const users = ref<any[]>([]);
    const loading = ref(true);
    const statusMessage = ref("");
    const errorMessage = ref("");
    const updatingId = ref<string | null>(null);
    const deletingId = ref<string | null>(null);

    const adminKey = (window as any).__ADMIN_API_KEY || (window as any).ADMIN_API_KEY || "";

    async function fetchUsers() {
      loading.value = true;
      statusMessage.value = "";
      errorMessage.value = "";
      try {
        const headers: any = {};
        if (adminKey) headers["Authorization"] = "Bearer " + adminKey;
        const res = await fetch("/api/admin/users", { headers });
        if (!res.ok) throw new Error(res.statusText);
        users.value = await res.json();
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal memuat pengguna";
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

    async function changeRole(u: any) {
      statusMessage.value = "";
      errorMessage.value = "";
      updatingId.value = u.id;
      try {
        const headers: any = { "Content-Type": "application/json" };
        if (adminKey) headers["Authorization"] = "Bearer " + adminKey;
        const res = await fetch("/api/admin/users/" + u.id, {
          method: "PUT",
          headers,
          body: JSON.stringify({ role: u.role }),
        });
        if (!res.ok) throw new Error("update failed");
        await res.json();
        statusMessage.value = "Role diperbarui";
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal memperbarui role";
      } finally {
        updatingId.value = null;
      }
    }

    async function confirmDelete(u: any) {
      if (!confirm("Hapus user " + u.email + " ?")) return;
      statusMessage.value = "";
      errorMessage.value = "";
      deletingId.value = u.id;
      try {
        const headers: any = {};
        if (adminKey) headers["Authorization"] = "Bearer " + adminKey;
        const res = await fetch("/api/admin/users/" + u.id, { method: "DELETE", headers });
        if (res.status !== 204) throw new Error("delete failed");
        users.value = users.value.filter((x) => x.id !== u.id);
        statusMessage.value = "User dihapus";
      } catch (err: any) {
        console.error(err);
        errorMessage.value = err.message || "Gagal menghapus user";
      } finally {
        deletingId.value = null;
      }
    }

    onMounted(() => {
      fetchUsers();
    });
    return { users, loading, formatDate, changeRole, confirmDelete, statusMessage, errorMessage, updatingId, deletingId };
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

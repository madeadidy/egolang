<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import Card from "@/components/ui/Card.vue";
import CardHeader from "@/components/ui/CardHeader.vue";
import CardTitle from "@/components/ui/CardTitle.vue";
import CardContent from "@/components/ui/CardContent.vue";
import Button from "@/components/ui/Button.vue";
import { Download } from "lucide-vue-next";

const orders = ref([] as any[]);
const start = ref("");
const end = ref("");
const customer = ref("");
const statusFilter = ref("");
const paymentStatus = ref("");
const page = ref(1);
const perPage = ref(20);
const totalCount = ref(0);
const totalRevenue = ref("0");
const revenueByStatus = ref({} as Record<string, string>);

const csvUrl = computed(() => {
  // use the API export endpoint so filters match the current report table
  const parts: string[] = [];
  if (start.value) parts.push(`start=${start.value}`);
  if (end.value) parts.push(`end=${end.value}`);
  if (customer.value) parts.push(`customer=${encodeURIComponent(customer.value)}`);
  if (statusFilter.value) parts.push(`status=${encodeURIComponent(statusFilter.value)}`);
  if (paymentStatus.value) parts.push(`payment_status=${encodeURIComponent(paymentStatus.value)}`);
  parts.push(`export=csv`);
  const q = parts.length ? `?${parts.join("&")}` : "";
  return `/api/admin/reports/transactions${q}`;
});

async function fetchReport() {
  const params: string[] = [];
  if (start.value) params.push(`start=${start.value}`);
  if (end.value) params.push(`end=${end.value}`);
  if (customer.value) params.push(`customer=${encodeURIComponent(customer.value)}`);
  if (statusFilter.value) params.push(`status=${encodeURIComponent(statusFilter.value)}`);
  if (paymentStatus.value) params.push(`payment_status=${encodeURIComponent(paymentStatus.value)}`);
  params.push(`page=${page.value}`);
  params.push(`per_page=${perPage.value}`);
  const q = params.length ? "?" + params.join("&") : "";
  const headers: any = { "Content-Type": "application/json" };
  const adminKey = (window as any).__ADMIN_API_KEY || (window as any).ADMIN_API_KEY || "";
  if (adminKey) headers["Authorization"] = "Bearer " + adminKey;
  const res = await fetch("/api/admin/reports/transactions" + q, { headers });
  if (res.ok) {
    const body = await res.json();
    orders.value = body.orders || [];
    if (body.meta) {
      totalCount.value = body.meta.total_count || 0;
      totalRevenue.value = body.meta.total_revenue || "0";
      revenueByStatus.value = body.meta.revenue_by_status || {};
    }
  } else {
    orders.value = [];
    totalCount.value = 0;
    totalRevenue.value = "0";
  }
}

function prevPage() {
  if (page.value > 1) {
    page.value--;
    fetchReport();
  }
}
function nextPage() {
  const maxPage = Math.ceil(totalCount.value / perPage.value);
  if (page.value < maxPage) {
    page.value++;
    fetchReport();
  }
}

onMounted(() => {
  fetchReport();
});
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight">Transactions Report</h1>
        <p class="text-muted-foreground">Filter orders and export transaction CSV</p>
      </div>
      <div class="flex items-center gap-2">
        <input v-model="start" type="date" class="input" />
        <input v-model="end" type="date" class="input" />
        <input v-model="customer" type="text" placeholder="Customer name or email" class="input" />
        <select v-model="statusFilter" class="input">
          <option value="">All Status</option>
          <option value="pending">Pending</option>
          <option value="received">Received</option>
          <option value="delivered">Delivered</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <select v-model="paymentStatus" class="input">
          <option value="">All Payments</option>
          <option value="paid">Paid</option>
          <option value="unpaid">Unpaid</option>
        </select>
        <Button variant="outline" size="sm" @click="fetchReport">
          <Download class="mr-2 h-4 w-4" />
          Refresh
        </Button>
        <a :href="csvUrl" class="inline-block">
          <Button variant="ghost" size="sm">
            <Download class="mr-2 h-4 w-4" />
            Export CSV
          </Button>
        </a>
      </div>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Transactions</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="mb-4 flex gap-4">
          <div class="p-3 bg-gray-50 rounded">
            <div class="text-sm text-muted-foreground">Total Orders</div>
            <div class="text-lg font-bold">{{ totalCount }}</div>
          </div>
          <div class="p-3 bg-gray-50 rounded">
            <div class="text-sm text-muted-foreground">Total Revenue</div>
            <div class="text-lg font-bold">{{ totalRevenue }}</div>
          </div>
          <div class="flex gap-2 items-stretch">
            <div v-for="(val, key) in revenueByStatus" :key="key" class="p-3 bg-gray-50 rounded">
              <div class="text-sm text-muted-foreground">{{ key }}</div>
              <div class="text-lg font-bold">{{ val }}</div>
            </div>
          </div>
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full table-auto">
            <thead>
              <tr>
                <th class="text-left">Order</th>
                <th class="text-left">Date</th>
                <th class="text-left">Customer</th>
                <th class="text-right">Items</th>
                <th class="text-right">Total</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="o in orders" :key="o.id" class="border-t">
                <td>{{ o.code }}</td>
                <td>{{ new Date(o.order_date).toLocaleString() }}</td>
                <td>{{ o.user_email || (o.user && o.user.first_name + " " + o.user.last_name) }}</td>
                <td class="text-right">{{ o.items_count }}</td>
                <td class="text-right">{{ o.grand_total }}</td>
              </tr>
              <tr v-if="orders.length === 0">
                <td colspan="5" class="text-center py-4 text-muted-foreground">No transactions</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>

    <div class="flex items-center justify-between">
      <div class="space-x-2">
        <button class="btn" @click="prevPage">Prev</button>
        <button class="btn" @click="nextPage">Next</button>
      </div>
      <div>
        Page {{ page }} — per page:
        <select v-model.number="perPage" @change="fetchReport">
          <option :value="10">10</option>
          <option :value="20">20</option>
          <option :value="50">50</option>
        </select>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
// Reactive metrics populated from server APIs
const totalTransactions = ref<number | string>(0)
const totalUsers = ref<number | string>(0)
const totalProducts = ref<number | string>(0)
const totalRevenue = ref<number | string>(0)

function adminHeaders() {
  const h: Record<string, string> = { 'Content-Type': 'application/json' }
  const k = (window as any).__ADMIN_API_KEY || (window as any).ADMIN_API_KEY || ''
  if (k) h['Authorization'] = 'Bearer ' + k
  return h
}

async function loadMetrics() {
  try {
    // transactions summary (meta.total_count, meta.total_revenue)
    const resR = await fetch('/api/admin/reports/transactions', { headers: adminHeaders() })
    if (resR.ok) {
      const j = await resR.json()
      totalTransactions.value = j.meta?.total_count ?? 0
      totalRevenue.value = j.meta?.total_revenue ?? 0
    }

    // users count
    const resU = await fetch('/api/admin/users', { headers: adminHeaders() })
    if (resU.ok) {
      const users = await resU.json()
      totalUsers.value = Array.isArray(users) ? users.length : 0
    }

    // products count
    const resP = await fetch('/api/admin/products', { headers: adminHeaders() })
    if (resP.ok) {
      const products = await resP.json()
      totalProducts.value = Array.isArray(products) ? products.length : 0
    }
  } catch (err) {
    console.error('Failed to load dashboard metrics', err)
  }
}

onMounted(() => {
  loadMetrics()
})

function formatIDR(value: number | string) {
  if (value === null || value === undefined) return 'Rp 0'
  // Accept numeric or numeric string
  const cleaned = String(value).replace(/[^0-9\-\.]/g, '')
  const num = Number(cleaned)
  if (isNaN(num) || !isFinite(num)) return 'Rp 0'
  const intPart = Math.floor(Math.abs(num))
  const sign = num < 0 ? '-' : ''
  const withDots = intPart.toString().replace(/\B(?=(\d{3})+(?!\d))/g, '.')
  return `${sign}Rp ${withDots}`
}
import Card from "@/components/ui/Card.vue";
import CardHeader from "@/components/ui/CardHeader.vue";
import CardTitle from "@/components/ui/CardTitle.vue";
import CardContent from "@/components/ui/CardContent.vue";
import { Users, Building2, TrendingUp, DollarSign } from "lucide-vue-next";
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-xl font-bold tracking-tight">Dashboard</h1>
      <p class="text-muted-foreground">Overview transaksi dan metrik utama toko</p>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Total Transaksi</CardTitle>
          <Users class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ totalTransactions }}</div>
          <p class="text-xs text-muted-foreground">&nbsp;</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Jumlah Users</CardTitle>
          <Building2 class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ totalUsers }}</div>
          <p class="text-xs text-muted-foreground">&nbsp;</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Jumlah Products</CardTitle>
          <TrendingUp class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ totalProducts }}</div>
          <p class="text-xs text-muted-foreground">&nbsp;</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">Total Pendapatan</CardTitle>
          <DollarSign class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ formatIDR(totalRevenue) }}</div>
          <p class="text-xs text-muted-foreground">&nbsp;</p>
        </CardContent>
      </Card>
    </div>

    <!-- Revenue Trend removed -->

    <!-- Recent Activity and Upcoming Tasks removed as requested -->
  </div>
</template>

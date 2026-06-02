import { createRouter, createWebHistory } from "vue-router";
import MainLayout from "@/layouts/MainLayout.vue";

const router = createRouter({
  history: createWebHistory("/material-dashboard-shadcn-vue/"),
  routes: [
    {
      path: "/",
      component: MainLayout,
      children: [
        {
          path: "",
          redirect: "/dashboard",
        },
        {
          path: "dashboard",
          name: "Dashboard",
          component: () => import("@/views/Dashboard.vue"),
        },

        {
          path: "reports",
          name: "Reports",
          component: () => import("@/views/Reports.vue"),
        },
        {
          path: "products",
          name: "Products",
          component: () => import("@/views/Products.vue"),
        },
        {
          path: "orders",
          name: "Orders",
          component: () => import("@/views/Orders.vue"),
        },
        {
          path: "customers",
          name: "Customers",
          component: () => import("@/views/Customers.vue"),
        },
        {
          path: "users",
          name: "Users",
          component: () => import("@/views/Users.vue"),
        },

        {
          path: "settings",
          name: "Settings",
          component: () => import("@/views/Settings.vue"),
        },
        {
          path: "docs",
          name: "Docs",
          component: () => import("@/views/Docs.vue"),
        },
      ],
    },
  ],
});

export default router;

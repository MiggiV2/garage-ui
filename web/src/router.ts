import { createRouter, createWebHistory } from "vue-router";
import OverviewView from "./views/OverviewView.vue";
import BucketsView from "./views/BucketsView.vue";
import BucketDetailView from "./views/BucketDetailView.vue";
import KeysView from "./views/KeysView.vue";
import KeyDetailView from "./views/KeyDetailView.vue";

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: OverviewView, meta: { title: "Overview" } },
    { path: "/buckets", component: BucketsView, meta: { title: "Buckets" } },
    { path: "/buckets/:id", component: BucketDetailView, props: true, meta: { title: "Bucket" } },
    { path: "/keys", component: KeysView, meta: { title: "Access keys" } },
    { path: "/keys/:id", component: KeyDetailView, props: true, meta: { title: "Access key" } },
    { path: "/:pathMatch(.*)*", redirect: "/" },
  ],
});

router.afterEach((to) => {
  document.title = `${to.meta.title ?? ""} · Garage Console`;
});

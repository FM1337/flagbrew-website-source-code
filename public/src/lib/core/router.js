import Vue from "vue";
import VueRouter from "vue-router";
import store from "~/lib/core/state.js";

import { AdminLogs, AdminSettings, GPSS, Index, Pokemon, Privacy, Project, Auth, NotFound } from "~/pages";
Vue.use(VueRouter);

const routes = [
  { name: Index.name, path: "/", component: Index, meta: { title: "Index" } },
  { name: Auth.name, path: "/ui/auth/github/callback", component: Auth },
  { name: Project.name, path: "/projects/:project", component: Project, meta: { title: Project.name } },
  { name: AdminSettings.name, path: "/admin/settings", component: AdminSettings },
  { name: AdminLogs.name, path: "/admin/logs", component: AdminLogs },
  { name: GPSS.name, path: "/gpss", component: GPSS, meta: { title: "GPSS" } },
  { name: Pokemon.name, path: "/gpss/:pokemon", component: Pokemon, meta: { title: "Placeholder" } },
  { name: Privacy.name, path: "/privacy", component: Privacy, meta: { title: "Privacy Policy" } },
  { name: "catchall", path: "*", redirect: "/404" },
  { name: NotFound.name, path: "/404", component: NotFound, meta: { title: "Page not found" } },
];

const router = new VueRouter({ routes, mode: "history" });
router.beforeEach((to, from, next) => {
  if (to.meta.title !== undefined) {
    document.title = `${to.meta.title} · FlagBrew`;
  } else {
    document.title = "FlagBrew";
  }

  if (from.name == null && !to.name.startsWith("auth-")) {
    store
      .dispatch("get_auth")
      .catch(() => {
        if (!!to.meta.auth_required) {
          window.location.replace("/api/v2/auth/github/redirect");
        }
      })
      .finally(() => {
        next();
      });
  } else {
    if (!store.state.auth) {
      // Assume we already fetched auth earlier.
      if (!!to.meta.auth_required) {
        window.location.replace("/api/v2/auth/github/redirect");
      }
    }
    next();
  }
});

export default router;

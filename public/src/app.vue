<template>
  <v-app>
    <vue-progress-bar></vue-progress-bar>
    <core-navbar v-on:toggle-sidebar="toggleSideBar"></core-navbar>
    <core-sidebar :toggled="showSidebar" @update-toggled="toggleSideBar"></core-sidebar>
    <v-main>
      <v-container
        fluid
        fill-height
        :class="`${$vuetify.theme.dark ? '' : 'grey'} lighten-4 d-flex flex-column align-stretch justify-start pa-0`"
      >
        <router-view></router-view>
      </v-container>
    </v-main>
    <core-snack></core-snack>
    <core-footer></core-footer>
  </v-app>
</template>

<script>
import CoreNavbar from "~/components/core/navbar";
import CoreSidebar from "~/components/core/sidebar";
import CoreFooter from "~/components/core/footer";
import CoreSnack from "~/components/core/snack";

export default {
  name: "app",
  components: { CoreNavbar, CoreSidebar, CoreFooter, CoreSnack },
  data: function() {
    return {
      showSidebar: this.$vuetify.breakpoint.mdAndUp,
    };
  },
  created: function() {
    const vm = this;
    vm.$Progress.start();
    vm.$router.beforeEach((to, from, next) => {
      if (to.meta.progress !== undefined) {
        let meta = to.meta.progress;
        vm.$Progress.parseMeta(meta);
      }

      vm.$Progress.start();
      next();
    });

    this.$router.afterEach((to, from) => {
      this.$Progress.finish();
    });

    this.$api.interceptors.request.use(
      (config) => {
        this.$Progress.start();
        return config;
      },
      (error) => {
        this.$Progress.fail();
        return Promise.reject(error);
      }
    );

    this.$api.interceptors.response.use(
      (response) => {
        this.$Progress.finish();
        return response;
      },
      (error) => {
        this.$Progress.fail();

        // interceptor for redirecting on 401 with authenticated=false.
        if (error.response.status == 401 && !error.response.data.authenticated) {
          // user has lost authentication status, redirect them to the home page.
          if (!this.$route.name.startsWith("auth-") && this.$route.name !== "index") {
            this.$store.dispatch("logout");
            this.$router.push({ name: "index" });
          }
        }
        return Promise.reject(error);
      }
    );
    this.$store.dispatch("fetch_projects");
    this.$store.dispatch("fetch_members");
  },
  mounted: function() {
    this.$Progress.finish();
  },
  methods: {
    toggleSideBar(state = null) {
      const vm = this;
      if (state !== null) {
        vm.showSidebar = state;
      } else {
        vm.showSidebar = !vm.showSidebar;
      }
    },
  },
};
</script>

<style>
.clickable {
  cursor: pointer;
}
.v-snack [role="progressbar"] {
  display: none;
}

::-webkit-scrollbar {
  width: 10px;
  height: 6px;
}
::-webkit-scrollbar-track-piece {
  background-color: #f5f5f5;
  background-clip: padding-box;
}
::-webkit-scrollbar-thumb {
  background-color: #1678c2;
  background-clip: padding-box;
  border: 2px solid #ffffff;
  border-radius: 6px;
}
::-webkit-scrollbar-thumb:window-inactive {
  background-color: #1678c2;
}
</style>

<template>
  <v-navigation-drawer v-model="drawer" @input="updateToggledStatus" app :expand-on-hover="$vuetify.breakpoint.mdAndUp" hide-overlay>
    <v-divider class="pb-15"></v-divider>
    <!-- <v-list-item>
      <v-list-item-content>
        <v-list-item-title class="title">Flagbrew</v-list-item-title>
      </v-list-item-content>
    </v-list-item> -->
    <v-list dense>
      <v-list-item v-for="item in items" :key="item.title" link @click="navigate(item.page, item.enabled)">
        <v-list-item-icon>
          <v-icon>{{ item.icon }}</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
      <v-list-group no-action>
        <template v-slot:activator>
          <v-list-item-icon>
            <v-icon>mdi-github</v-icon>
          </v-list-item-icon>

          <v-list-item-content>
            <v-list-item-title>Projects</v-list-item-title>
          </v-list-item-content>
        </template>
        <v-list-item v-for="project in projects" :key="project.repo_id" link @click="navigate('/projects/' + project.name, true)">
          <v-list-item-content>
            <v-list-item-title v-text="project.name"></v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list-group>
      <v-divider></v-divider>
      <v-list-item link @click="navigate('/privacy', true)">
        <v-list-item-icon>
          <v-icon>mdi-book</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>Privacy Policy</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
      <v-list-item>
        <v-list-item-content>
          <v-switch v-model="darkMode" prepend-icon="mdi-moon-waning-crescent" label="Dark Mode"></v-switch>
        </v-list-item-content>
      </v-list-item>
    </v-list>
  </v-navigation-drawer>
</template>

<script>
export default {
  name: "core-sidebar",
  data() {
    return {
      items: [
        { title: "Home", icon: "dashboard", page: "/", enabled: true },
        { title: "GPSS", icon: "mdi-earth", page: "/gpss", enabled: true },
      ],
      drawer: false,
      darkMode: false,
    };
  },
  created() {
    this.drawer = this.toggled;
    this.darkMode = this.$vuetify.theme.dark;
  },
  props: {
    toggled: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    projects() {
      return this.$store.state.projects;
    },
  },
  watch: {
    toggled: function(state) {
      const vm = this;
      if (state !== vm.drawer) {
        vm.drawer = state;
      }
    },
    darkMode: function(state) {
      const vm = this;
      localStorage.setItem("dark", state);
      vm.$vuetify.theme.dark = state;
    },
  },
  methods: {
    navigate: function(page, enabled) {
      const vm = this;
      if (vm.$router.history.current.fullPath !== page && enabled) {
        vm.$router.push(page);
      }
    },
    updateToggledStatus(state) {
      const vm = this;
      if (state !== vm.toggled) {
        vm.$emit("update-toggled", state);
      }
    },
  },
};
</script>

<style scoped>
.v-navigation-drawer {
  z-index: 2;
}
</style>

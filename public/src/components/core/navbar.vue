<template>
  <v-app-bar :color="`purple darken-${$vuetify.theme.dark ? '3' : '1'}`" app clipped-left>
    <v-app-bar-nav-icon @click="$emit('toggle-sidebar')"></v-app-bar-nav-icon>
    <v-icon class="hidden-sm-and-down">flag</v-icon>
    <router-link custom v-slot="{ navigate }" :to="{ name: 'index' }">
      <span class="title ml-3 mr-5 clickable" @click="navigate">
        <v-toolbar-title> Flag<span class="font-weight-light">Brew</span> </v-toolbar-title>
      </span>
    </router-link>
    <!-- <v-text-field solo text hide-details label="Search" prepend-inner-icon="search"></v-text-field> -->
    <v-spacer></v-spacer>

    <v-btn v-if="auth == false" text light href="/api/v2/auth/github/redirect">Login</v-btn>

    <v-progress-circular v-if="auth == null" indeterminate color="purple"></v-progress-circular>

    <!-- :nudge-width="200"; TODO: https://github.com/vuetifyjs/vuetify/issues/3438 -->
    <v-menu v-if="auth" :nudge-width="200" offset-y :nudge-top="-15">
      <template v-slot:activator="{ on, attrs }">
        <v-avatar v-on="on" v-bind="attrs" size="36px">
          <img class="user-avatar" :src="auth.avatar_url" :alt="auth.name" />
        </v-avatar>
      </template>

      <v-list>
        <v-list-item>
          <v-list-item-avatar>
            <img :src="auth.avatar_url" :alt="auth.name" />
          </v-list-item-avatar>

          <v-list-item-content>
            <v-list-item-title>{{ auth.name }}</v-list-item-title>
            <v-list-item-subtitle>{{ auth.username }}</v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </v-list>

      <v-divider></v-divider>

      <v-list>
        <v-list-item to="/admin/settings">
          <v-list-item-action>
            <v-icon>security</v-icon>
          </v-list-item-action>
          <v-list-item-title>Admin settings</v-list-item-title>
        </v-list-item>
        <v-list-item to="/admin/logs">
          <v-list-item-action>
            <v-icon>mdi-book-search</v-icon>
          </v-list-item-action>
          <v-list-item-title>Logs</v-list-item-title>
        </v-list-item>
        <v-list-item href="/api/v2/auth/github/manage">
          <v-list-item-action>
            <v-icon>settings</v-icon>
          </v-list-item-action>
          <v-list-item-title>Manage Github permissions</v-list-item-title>
        </v-list-item>
        <v-list-item @click="logout()">
          <v-list-item-action>
            <v-icon>lock</v-icon>
          </v-list-item-action>
          <v-list-item-title>Sign out</v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
  </v-app-bar>
</template>

<script>
export default {
  name: "core-navbar",
  computed: {
    auth() {
      return this.$store.state.auth;
    },
  },
  methods: {
    test: function() {
      this.$http.get("/api/v2/test");
    },
    logout: function() {
      this.$store.dispatch("logout").then(() => {
        this.$router.push("/");
      });
    },
  },
};
</script>

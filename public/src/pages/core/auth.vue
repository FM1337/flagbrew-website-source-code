<template>
  <v-layout justify-center align-center>
    <v-progress-circular v-if="!error" :size="90" :width="7" color="light-blue" indeterminate></v-progress-circular>
    <v-alert v-else color="error" icon="warning">{{ error }}</v-alert>
  </v-layout>
</template>

<script>
export default {
  name: "auth-github-callback",
  data: () => ({
    error: false,
  }),
  mounted: function() {
    // this.$store.commit('disable_auth', true);
    // Send the oauth response GET data to the backend. We're doing
    // this on the frontend, so that way if there is an error, we can
    // better show it to the user. If we do it all on the backend,
    // we'd have to figure out some way of notifying the frontend,
    // which is... messy.
    this.$api
      .get("/auth/github/callback", {
        params: {
          code: this.$route.query.code,
          state: this.$route.query.state,
        },
      })
      .then(() => {
        this.$store.dispatch("get_auth").then(() => {
          this.$router.push("/");
        });
      })
      .catch((error) => {
        this.$store.dispatch("get_auth");
        this.error = error.response.data.error || error;
        this.$store.dispatch("showSnackbar", { text: "You are not authorized to login", color: "red", timeout: 5000 });
      });
  },
};
</script>

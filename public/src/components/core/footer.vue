<template>
  <v-footer padless>
    <v-card-text class="py-2 text-center">
      {{ new Date().getFullYear() }} — <strong>FlagBrew &copy;</strong>
      <br v-if="$vuetify.breakpoint.xsOnly" />
      <v-btn v-for="app in apps" :key="app.name" class="mx-4" icon @click="footerLinks(app.name)">
        <v-icon size="24px">
          {{ app.icon }}
        </v-icon>
      </v-btn>
    </v-card-text>
  </v-footer>
</template>

<script>
export default {
  name: "core-footer",
  data: () => ({
    apps: [
      { icon: "mdi-discord", name: "discord" },
      { icon: "mdi-twitter", name: "twitter" },
      { icon: "mdi-patreon", name: "patreon" },
      { icon: "mdi-github", name: "github" },
    ],
  }),
  computed: {
    year: function() {
      var d = new Date();
      return d.getFullYear();
    },
  },
  methods: {
    footerLinks(app) {
      switch (app) {
        case "discord":
          window.open("https://discord.gg/4Rv8DTzGFa");
          break;
        case "twitter":
          window.open("https://twitter.com/FlagBrewOrg");
          break;
        case "patreon":
          window.open("https://www.patreon.com/FlagBrew");
          break;
        case "github":
          window.open("https://github.com/FlagBrew");
          break;
      }
    },
  },
};
</script>

<style scoped>
span {
  font-size: 14px;
}
</style>

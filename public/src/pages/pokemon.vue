<template>
  <v-container>
    <v-overlay :absolute="true" :value="pkmn === null"></v-overlay>
    <v-row v-if="pkmn !== null">
      <v-col cols="12">
        <h2>Viewing {{ gender }} {{ pkmn.nickname }}</h2>
        <v-btn small color="primary" @click="$router.push('/gpss')">
          Back to GPSS
        </v-btn>
        <v-btn small color="success" @click="download">
          Download
        </v-btn>
        <v-divider />
      </v-col>
      <v-col cols="12">
        <gpss-pokemon :pokemon="pkmn" :data="data" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import pokemon from "../components/gpss/pokemon";
import { downloadFromBase64 } from "../lib/gpss/download";
export default {
  name: "pokemon",
  data: () => ({
    pkmn: null,
    data: null,
    gender: null,
  }),
  components: { "gpss-pokemon": pokemon },
  created() {
    const vm = this;
    this.$api
      .get(`/gpss/view/${this.$route.params.pokemon}`)
      .then((resp) => {
        vm.pkmn = resp.data.pokemon.pokemon;
        vm.gender = vm.pkmn.gender !== "-" ? (vm.pkmn.gender === "M" ? "♂" : "♀") : "⚲";
        vm.data = resp.data.pokemon;
        vm.$delete(vm.data, "pokemon");
        document.title = vm.pkmn.nickname + " · GPSS · FlagBrew";
      })
      .catch((error) => {
        vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        vm.$router.push("/404").catch(() => {});
      });
  },
  methods: {
    download: function(event) {
      const vm = this;
      if (event.detail != 1) {
        return;
      }
      if (!vm.downloading) {
        vm.downloading = true;
        vm.$api
          .get(`/gpss/download/pokemon/${vm.data.download_code}`)
          .then((resp) => {
            downloadFromBase64(resp.data.pokemon, resp.data.filename);
            vm.data.current_downloads++;
            vm.data.lifetime_downloads++;
          })
          .catch((error) => {
            vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
          })
          .finally(() => {
            vm.downloading = false;
          });
      }
    },
  },
};
</script>

<style></style>

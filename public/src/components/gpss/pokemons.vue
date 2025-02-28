<template>
  <v-container fluid>
    <v-data-iterator :items="pokemon" :items-per-page="12" hide-default-footer>
      <template v-slot:default="props">
        <v-row>
          <v-col v-for="(pkmn, index) in props.items" :key="pkmn.download_code" cols="12" md="6" lg="4" xl="3">
            <v-card height="100%" width="100%" class="d-flex flex-column">
              <v-list-item class="equal-height-item">
                <v-list-item-avatar :size="$vuetify.breakpoint.mobile ? 50 : 100">
                  <v-img position="53% -21px" :src="pkmn.pokemon.sprites.species" :alt="pkmn.pokemon.species" />
                </v-list-item-avatar>
                <v-list-item-content>
                  <v-list-item-title class="headline">
                    <v-tooltip top>
                      <template v-slot:activator="{ on }">
                        <v-icon :color="pkmn.pokemon.is_legal ? 'green' : 'red'" v-on="on">{{
                          pkmn.pokemon.is_legal ? "mdi-check" : "mdi-close"
                        }}</v-icon>
                      </template>
                      <span>This Pokemon has {{ pkmn.pokemon.is_legal ? "passed" : "failed" }} legality checks!</span>
                    </v-tooltip>
                    <span class="text-truncate">
                      {{ pkmn.pokemon.nickname !== undefined ? pkmn.pokemon.nickname : pkmn.pokemon.species }}
                    </span>
                    <v-icon v-if="pkmn.pokemon.gender !== '-'">{{
                      pkmn.pokemon.gender === "F" ? "mdi-gender-female" : "mdi-gender-male"
                    }}</v-icon>
                    <span v-else>⚲</span>
                  </v-list-item-title>
                  <v-list-item-subtitle>{{ pkmn.pokemon.species }}</v-list-item-subtitle>
                  <v-list-item-subtitle>
                    <v-tooltip top v-if="pkmn.pokemon.is_shiny">
                      <template v-slot:activator="{ on }">
                        <v-icon color="orange" v-on="on">mdi-star</v-icon>
                      </template>
                      <span>This Pokemon is shiny!</span>
                    </v-tooltip>

                    <v-tooltip top v-if="pkmn.pokemon.is_egg">
                      <template v-slot:activator="{ on }">
                        <v-icon color="blue-grey lighten-2" v-on="on">mdi-egg</v-icon>
                      </template>
                      <span>This Pokemon is an egg!</span>
                    </v-tooltip>

                    <v-tooltip top v-if="pkmn.patreon">
                      <template v-slot:activator="{ on }">
                        <v-icon color="red" v-on="on">mdi-heart</v-icon>
                      </template>
                      <span>This Pokemon was uploaded by a Patreon!</span>
                    </v-tooltip>
                    <v-tooltip top v-if="!pkmn.approved">
                      <template v-slot:activator="{ on }">
                        <v-icon color="info" v-on="on">mdi-information-outline</v-icon>
                      </template>
                      <span>This Pokemon has not been approved yet!</span>
                    </v-tooltip>
                  </v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>

              <v-divider></v-divider>

              <v-list dense>
                <v-list-item>
                  <v-list-item-content>Ability:</v-list-item-content>
                  <v-list-item-content class="align-end">{{ pkmn.pokemon.ability ? pkmn.pokemon.ability : "N/A" }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>Level:</v-list-item-content>
                  <v-list-item-content class="align-end">{{ pkmn.pokemon.level }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>OT/ID:</v-list-item-content>
                  <v-list-item-content class="align-end">{{ pkmn.pokemon.ot }}/{{ pkmn.pokemon.tid }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>HT:</v-list-item-content>
                  <v-list-item-content class="align-end">{{
                    pkmn.pokemon.not_ot !== "" && pkmn.pokemon.not_ot !== undefined ? pkmn.pokemon.not_ot : "N/A"
                  }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>Generation:</v-list-item-content>
                  <v-list-item-content class="align-end">{{ pkmn.generation }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>Original Language:</v-list-item-content>
                  <v-list-item-content class="align-end">{{ pkmn.pokemon.ot_lang ? pkmn.pokemon.ot_lang : "N/A" }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>Original Game:</v-list-item-content>
                  <v-list-item-content class="align-end">{{ pkmn.pokemon.version ? pkmn.pokemon.version : "N/A" }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>Met Location:</v-list-item-content>
                  <v-list-item-content class="align-end">{{
                    pkmn.pokemon.met_data.name ? pkmn.pokemon.met_data.name : "N/A"
                  }}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content>Held Item:</v-list-item-content>
                  <v-img
                    v-if="pkmn.pokemon.held_item !== 'None'"
                    class="align-self-stretch"
                    max-width="15%"
                    :aspect-ratio="16 / 9"
                    :src="pkmn.pokemon.sprites.item"
                    :alt="pkmn.pokemon.species"
                  />
                  <v-list-item-content class="align-end">{{ pkmn.pokemon.held_item }}</v-list-item-content>
                </v-list-item>
              </v-list>
              <v-spacer v-if="pkmn.pokemon.met_loc !== undefined && pkmn.pokemon.met_loc.length < 14"></v-spacer>
              <v-divider></v-divider>
              <v-card-actions>
                <v-btn text color="deep-purple accent-4" :disabled="downloading" @click="download(index, $event)">
                  Download
                </v-btn>
                <v-btn text color="deep-purple accent-4" @click="view(pkmn.download_code)">
                  View
                </v-btn>
                <v-spacer></v-spacer>
                <span class="caption">{{ pkmn.current_downloads }}</span>
                <v-icon>mdi-download</v-icon>
                <span class="caption">{{ pkmn.lifetime_downloads }}</span>
                <v-icon>mdi-earth</v-icon>
              </v-card-actions>
            </v-card>
          </v-col>
        </v-row>
      </template>
    </v-data-iterator>
  </v-container>
</template>

<script>
import { downloadFromBase64 } from "../../lib/gpss/download";

export default {
  name: "gpss-pokemons",
  props: {
    page: {
      type: Number,
      default: 0,
    },
    searchQuery: {
      type: Object,
      default: {},
    },
    sort: {
      type: Object,
      required: true,
    },
    resetPage: {
      type: Boolean,
      default: false,
    },
  },
  created: function() {
    this.getData();
  },
  data: () => ({
    pages: 0,
    pokemon: [],
    loading: false,
    downloading: false,
  }),
  methods: {
    download: function(index, event) {
      const vm = this;
      if (event.detail != 1) {
        return;
      }
      if (!vm.downloading) {
        vm.downloading = true;
        vm.$api
          .get(`/gpss/download/pokemon/${vm.pokemon[index].download_code}`)
          .then((resp) => {
            downloadFromBase64(resp.data.pokemon, resp.data.filename);
            vm.pokemon[index].current_downloads++;
            vm.pokemon[index].lifetime_downloads++;
          })
          .catch((error) => {
            vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
          })
          .finally(() => {
            vm.downloading = false;
          });
      }
    },
    view: function(downloadCode) {
      this.$router.push(`/gpss/${downloadCode}`);
    },
    getData: function() {
      const vm = this;
      if (!vm.loading) {
        vm.loading = true;
        this.$api
          .post(`/gpss/search/pokemon?page=${this.page}`, {
            ...vm.searchQuery,
            ...vm.sort,
          })
          .then((resp) => {
            this.pages = resp.data.pages;
            this.pokemon = resp.data.pokemon;
            this.$emit("update-page-count", this.pages);
            this.loading = false;
          })
          .catch((error) => {
            this.loading = false;
            vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
          });
      }
    },
  },
  watch: {
    page: function() {
      const vm = this;
      if (!vm.resetPage) {
        vm.getData();
      }
    },
    searchQuery: function() {
      const vm = this;
      vm.getData();
    },
    sort: function() {
      const vm = this;
      vm.getData();
    },
    loading: function() {
      this.$emit("update-loading", this.loading);
    },
  },
};
</script>

<style scoped>
.equal-height-item {
  height: 96px;
}
.card-outter {
  position: relative;
  padding-bottom: 50px;
}
.card-actions {
  position: absolute;
  bottom: 0;
}
</style>

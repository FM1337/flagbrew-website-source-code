<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12" md="6" v-if="!pokemon.is_legal">
        <span class="text-h4">Notice</span>
        <p class="subtitle-2">This Pokemon has failed legality checks and likely would not be usable online!</p>
        <v-expansion-panels>
          <v-expansion-panel>
            <v-expansion-panel-header disable-icon-rotate>
              <div>
                <v-badge color="error" :content="legalityIssues.length">
                  Legality Issues
                </v-badge>
              </div>
            </v-expansion-panel-header>
            <v-expansion-panel-content>
              <p class="subtitle-2" v-for="(issue, i) in legalityIssues" :key="i">{{ issue }}</p>
            </v-expansion-panel-content>
          </v-expansion-panel>
        </v-expansion-panels>
      </v-col>
      <v-col cols="12" v-if="!pokemon.is_legal">
        <v-divider />
      </v-col>
      <v-col cols="12">
        <v-row>
          <v-col cols="12" md="6">
            <Info :pokemon="pokemon" />
          </v-col>
          <v-col cols="12" md="6">
            <Stats :stats="pokemon.stats ? pokemon.stats : []" />
          </v-col>
          <v-col cols="12" md="6">
            <Moves :moves="pokemon.moves ? pokemon.moves : []" :generation="pokemon.generation" />
          </v-col>
          <v-col cols="12" md="6">
            <Downloads :lifetimeDownloads="data.lifetime_downloads" :currentDownloads="data.current_downloads" />
          </v-col>
          <v-col cols="12" md="6">
            <Ribbons :ribbons="pokemon.ribbons ? pokemon.ribbons : []" />
          </v-col>
          <v-col cols="12" md="6">
            <Hex :base64="data.base_64" />
          </v-col>
        </v-row>
      </v-col>
      <v-col cols="12" v-if="$store.state.auth">
        <Moderation :downloadCode="data.download_code" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { Stats, Moves, Ribbons, Downloads, Info, Moderation, Hex } from "./pokemon-parts/";
export default {
  name: "gpss-pokemon",
  components: { Stats, Moves, Ribbons, Downloads, Info, Moderation, Hex },
  props: {
    pokemon: {
      type: Object,
      required: true,
    },
    data: {
      type: Object,
      required: true,
    },
  },
  computed: {
    legalityIssues: function() {
      return this.pokemon.illegal_reasons.split("\n");
    },
  },
};
</script>

<style></style>

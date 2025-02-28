<template>
  <v-card>
    <v-card-title>Moves</v-card-title>
    <v-divider></v-divider>
    <v-list class="transparent" two-line>
      <v-list-item v-for="move in pokemonMoves" :key="move.name">
        <v-list-item-icon>
          <v-icon :color="getColor(move.type)">
            {{ getIcon(move.type) }}
          </v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title>{{ move.name }}</v-list-item-title>
          <v-list-item-subtitle> {{ move.pp }} PP | {{ move.pp_ups }} PP Ups </v-list-item-subtitle>
        </v-list-item-content>
        <v-list-item-action>
          <v-tooltip right>
            <template v-slot:activator="{ on, attrs }">
              <v-btn icon @click="viewMoveInfo(move.name)" v-bind="attrs" v-on="on">
                <v-icon color="blue lighten-1">mdi-information</v-icon>
              </v-btn>
            </template>
            <span>View {{ move.name }} on Serebii</span>
          </v-tooltip>
        </v-list-item-action>
      </v-list-item>
    </v-list>
  </v-card>
</template>

<script>
export default {
  name: "gpss-pokemon-moves",
  props: {
    moves: {
      type: Array,
      required: true,
    },
    generation: {
      type: Number,
      required: true,
    },
  },
  methods: {
    viewMoveInfo(move) {
      let dex = "";
      switch (this.generation) {
        case 1:
          dex = "-rby";
          break;
        case 2:
          dex = "-gs";
          break;
        case 4:
          dex = "-dp";
          break;
        case 5:
          dex = "-bw";
          break;
        case 6:
          dex = "-xy";
          break;
        case 7:
          dex = "-sm";
          break;
        case 8:
          dex = "-swsh";
          break;
        case 9:
          dex = "-sv";
          break;
      }

      const moveName = move.toLowerCase().replace(" ", "");

      const url = `https://serebii.net/attackdex${dex}/${moveName}.shtml`;
      window.open(url);
    },
    getIcon(type) {
      switch (type) {
        case "Fire":
          return "mdi-fire";
        case "Ghost":
          return "mdi-ghost";
        case "Poison":
          return "mdi-skull-crossbones";
        case "Psychic":
          return "mdi-waveform";
        case "Grass":
          return "mdi-grass";
        case "Ground":
          return "mdi-earth";
        case "Ice":
          return "mdi-ice-cream";
        case "Rock":
          return "mdi-diamond-stone";
        case "Dragon":
          return "mdi-google-downasaur";
        case "Water":
          return "mdi-water";
        case "Bug":
          return "mdi-bug";
        case "Dark":
          return "mdi-moon-waning-crescent";
        case "Fighting":
          return "mdi-karate";
        case "Steel":
          return "mdi-pipe";
        case "Flying":
          return "mdi-bird";
        case "Electric":
          return "mdi-flash";
        case "Fairy":
          return "mdi-unicorn";
        case "Normal":
        default:
          return "mdi-file";
      }
    },
    getColor(type) {
      switch (type) {
        case "Fire":
          return "red";
        case "Ghost":
          return "indigo darken-4";
        case "Poison":
          return "deep-purple darken-1";
        case "Psychic":
          return "purple accent-3";
        case "Grass":
          return "light-green accent-3";
        case "Ground":
          return "brown lighten-1";
        case "Ice":
          return "light-blue lighten-4";
        case "Rock":
          return "lime darken-3";
        case "Dragon":
          return "deep-purple lighten-1";
        case "Water":
          return "light-blue accent-2";
        case "Bug":
          return "lime lighten-1";
        case "Dark":
          return "grey darken-4";
        case "Fighting":
          return "deep-orange darken-3";
        case "Steel":
          return "blue-grey darken-3";
        case "Flying":
          return "cyan accent-1";
        case "Electric":
          return "yellow accent-3";
        case "Fairy":
          return "purple accent-1";
        case "Normal":
        default:
          return "lime darken-2";
      }
    },
  },
  computed: {
    pokemonMoves: function() {
      return this.moves.filter((move) => move.name !== "None");
    },
  },
};
</script>

<style></style>

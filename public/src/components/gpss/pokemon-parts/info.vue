<template>
  <v-card>
    <v-card-title>Info</v-card-title>
    <v-expansion-panels multiple>
      <v-expansion-panel>
        <v-expansion-panel-header>
          Basic Info
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Name</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.nickname }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Species</v-list-item-title>
            </v-list-item-content>
            <v-list-item-avatar>
              <v-img position="50% -9px" :src="pokemon.sprites.species"></v-img>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title
                >{{ pokemon.species }}
                <!-- <v-btn icon>
                  <v-icon color="blue lighten-1">mdi-information</v-icon>
                </v-btn> -->
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Level</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.level }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Ability</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title
                >{{ pokemon.ability ? pokemon.ability : "N/A" }}
                <!-- <v-btn icon v-if="pokemon.ability">
                  <v-icon color="blue lighten-1">mdi-information</v-icon>
                </v-btn> -->
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Nature</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.nature ? pokemon.nature : "N/A" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>

          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Hidden Power</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.hp_type }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>Held Item:</v-list-item-content>
            <v-img
              v-if="pokemon.held_item !== 'None'"
              class="align-self-stretch"
              max-width="5%"
              :aspect-ratio="16 / 9"
              :src="pokemon.sprites.item"
              :alt="pokemon.held_item"
            />
            <v-list-item-content class="align-end">{{ pokemon.held_item }}</v-list-item-content>
          </v-list-item>
        </v-expansion-panel-content>
      </v-expansion-panel>
      <v-expansion-panel>
        <v-expansion-panel-header>
          Location Info
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Game Version</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.version ? pokemon.version : "N/A" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Met Location</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.met_data.name ? pokemon.met_data.name : "N/A" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Met Date</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title v-if="pokemon.met_data.year !== 1">{{
                `${pokemon.met_data.year}-${
                  String(pokemon.met_data.month).length === 1 ? `0${pokemon.met_data.month}` : pokemon.met_data.month
                }-${String(pokemon.met_data.day).length === 1 ? `0${pokemon.met_data.day}` : pokemon.met_data.day}` | date
              }}</v-list-item-title>
              <v-list-item-title v-else>N/A</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-expansion-panel-content>
      </v-expansion-panel>
      <v-expansion-panel>
        <v-expansion-panel-header>
          Egg Info
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Egg Location</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.egg_data.name ? pokemon.egg_data.name : "N/A" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Receive Date</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title v-if="pokemon.egg_data.year !== 1">{{
                `${pokemon.egg_data.year}-${
                  String(pokemon.egg_data.month).length === 1 ? `0${pokemon.egg_data.month}` : pokemon.egg_data.month
                }-${String(pokemon.egg_data.day).length === 1 ? `0${pokemon.egg_data.day}` : pokemon.egg_data.day}` | date
              }}</v-list-item-title>
              <v-list-item-title v-else>N/A</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-expansion-panel-content>
      </v-expansion-panel>
      <v-expansion-panel>
        <v-expansion-panel-header>
          Trainer Info
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>OT Name/ID</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.ot ? pokemon.ot : "N/A" }}/{{ pokemon.tid ? pokemon.tid : "N/A" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>OT Language</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.ot_lang ? pokemon.ot_lang : "N/A" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>HT Name</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ pokemon.ht ? pokemon.ht : "N/A" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
  </v-card>
</template>

<script>
export default {
  name: "gpss-pokemon-info",
  props: {
    pokemon: {
      type: Object,
      required: true,
    },
  },
};
</script>

<style></style>

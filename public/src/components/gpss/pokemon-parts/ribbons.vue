<template>
  <v-card>
    <v-card-title>Ribbons</v-card-title>
    <v-divider></v-divider>
    <v-card-text v-if="ribbons.length === 0">This Pokemon has no ribbons.</v-card-text>
    <v-list-item v-else>
      <v-list-item-content>
        <v-list-item-title v-for="(ribbon, i) in ribbons" :key="i">
          {{ ribbon }}
        </v-list-item-title>
      </v-list-item-content>
    </v-list-item>
  </v-card>
</template>

<script>
export default {
  name: "gpss-pokemon-ribbons",
  props: {
    ribbons: {
      type: Array,
      required: true,
    },
  },
};
</script>

<style></style>

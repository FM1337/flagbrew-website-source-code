<template>
  <v-card>
    <v-card-title>Stats</v-card-title>
    <v-divider></v-divider>
    <v-simple-table>
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">
              Stat
            </th>
            <th class="text-left">
              IV
            </th>
            <th class="text-left">
              EV
            </th>
            <th class="text-left">
              Total
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="stat in stats" :key="stat.stat_name">
            <td>{{ stat.name }}</td>
            <td>{{ stat.iv }}</td>
            <td>{{ stat.ev }}</td>
            <td>{{ stat.total }}</td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>
  </v-card>
</template>

<script>
export default {
  name: "gpss-pokemon-stats",
  props: {
    stats: {
      type: Array,
      required: true,
    },
  },
};
</script>

<style></style>

<template>
  <v-card>
    <v-card-title>Download Stats</v-card-title>
    <v-divider></v-divider>
    <download-chart v-if="lifetimeDownloads !== 0 && loaded" :chartData="chartData" :options="options"></download-chart>
    <v-card-text v-else>This Pokemon has no downloads yet.</v-card-text>
  </v-card>
</template>

<script>
import DownloadChart from "./download-chart.vue";
export default {
  name: "gpss-pokemon-downloads",
  components: { DownloadChart },
  props: {
    lifetimeDownloads: {
      type: Number,
      required: true,
    },
    currentDownloads: {
      type: Number,
      required: true,
    },
  },
  data: () => ({
    chartData: null,
    options: {
      responsive: true,
      maintainAspectRatio: false,
    },
    loaded: false,
  }),
  created() {
    if (this.lifetimeDownloads !== 0) {
      this.fillChartData();
    }
  },

  methods: {
    fillChartData() {
      const vm = this;
      vm.chartData = {
        labels: ["Current Downloads", "Lifetime Downloads"],
        datasets: [
          {
            label: "Download Stats",
            backgroundColor: ["#E46651", "#00D8FF"],
            data: [vm.currentDownloads, vm.lifetimeDownloads],
          },
        ],
      };
      vm.loaded = true;
    },
  },
  watch: {
    lifetimeDownloads: function() {
      this.fillChartData();
    },
  },
};
</script>

<style></style>

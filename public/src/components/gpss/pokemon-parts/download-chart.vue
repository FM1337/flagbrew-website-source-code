<script>
import { Pie } from "vue-chartjs";

export default {
  extends: Pie,
  name: "gpss-download-chart",
  props: {
    chartData: {
      type: Object,
      default: null,
    },
    options: {
      type: Object,
      default: null,
    },
  },
  mounted() {
    this.renderChart(this.chartData, this.options);
  },
  watch: {
    chartData: function() {
      this.$data._chart.destroy();
      this.renderChart(this.chartData, this.options);
    },
  },
};
</script>

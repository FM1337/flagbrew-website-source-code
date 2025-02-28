<template>
  <v-card>
    <v-card-title>Hex Data</v-card-title>
    <v-card-text>
      <pre><code>{{ hex }}</code></pre>
    </v-card-text>
  </v-card>
</template>

<script>
export default {
  name: "gpss-pokemon-hex",
  props: {
    base64: {
      type: String,
      required: true,
    },
  },
  data: () => ({
    hex: null,
  }),

  created() {
    const vm = this;
    let raw = atob(vm.base64);
    let result = "";
    for (let i = 0; i < raw.length; i++) {
      const hex = raw.charCodeAt(i).toString(16);
      result += hex.length === 2 ? hex : "0" + hex;
    }

    let ret = [];

    for (let i = 0, len = result.length; i < len; i += 2) {
      ret.push(result.substr(i, 2));
    }

    vm.hex = ret.join(" ");
  },
};
</script>

<style scoped>
pre {
  padding: 5px;
  margin: 5px;
  white-space: pre-wrap;
  white-space: -moz-pre-wrap;
  white-space: -pre-wrap;
  white-space: -o-pre-wrap;
  word-wrap: break-word;
}
</style>

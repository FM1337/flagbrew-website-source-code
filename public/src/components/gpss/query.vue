<template>
  <v-row>
    <v-col cols="11" v-if="componentName === 'slider-range'">
      <v-range-slider v-bind="componentProps" @change="sliderRangeUpdate" v-model="localValue" :thumb-label="true"></v-range-slider>
    </v-col>
    <v-col v-else cols="11">
      <v-row>
        <v-col cols="12" md="4">
          <v-subheader>{{ name }}</v-subheader>
        </v-col>
        <v-col cols="12" md="3">
          <v-select :items="operators" v-model="operator" label="Operator" solo dense hint="The operator to use"></v-select>
        </v-col>
        <v-col cols="12" md="5"><component :is="componentToUse" v-bind="componentProps" v-model="localValue" /> </v-col>
      </v-row>
    </v-col>
    <v-col sm="1" cols="12">
      <v-tooltip top>
        <template v-slot:activator="{ on, attrs }">
          <v-btn icon v-bind="attrs" v-on="on" @click="deleteQuery">
            <v-icon>mdi-delete</v-icon>
          </v-btn>
        </template>
        <span>Delete Query</span>
      </v-tooltip>
    </v-col>
  </v-row>
</template>

<script>
export default {
  name: "gpss-query",
  data: () => ({
    operator: null,
  }),
  props: {
    name: {
      type: String,
    },
    operators: {
      type: Array,
    },
    default: {
      type: [Array, String, Number, Function, Boolean],
    },
    defaultOperator: {
      type: String,
    },
    componentName: {
      type: String,
      validator: function(value) {
        return ["slider-range", "select", "text", "autoComplete"];
      },
    },
    componentProps: {
      type: Object,
    },
  },
  created() {
    const vm = this;
    vm.value = vm.default;
    vm.localValue = vm.default;
    vm.operator = vm.defaultOperator;
  },
  computed: {
    localValue: {
      get() {
        return this.value;
      },
      set(value) {
        this.$emit("input", value);
      },
    },
    componentToUse: function() {
      const vm = this;
      const component = vm.componentName;
      switch (component) {
        case "select": {
          return "VSelect";
        }
        case "autoComplete": {
          return "VAutocomplete";
        }
        case "text": {
          return "VTextField";
        }
      }
    },
  },
  watch: {
    operator() {
      this.$emit("operatorChange", this.operator);
    },
  },
  methods: {
    deleteQuery() {
      this.$emit("delete", this.name);
    },
    sliderRangeUpdate(v) {
      const vm = this;
      vm.localValue = v;
    },
  },
};
</script>

<style></style>

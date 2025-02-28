<template>
  <v-container fluid class="mt-4">
    <v-container fluid class="d-flex">
      <v-text-field
        label="Query"
        v-model="basicQuery"
        solo
        :prepend-inner-icon="loading ? 'mdi-loading mdi-spin' : 'mdi-magnify'"
        hide-details
        v-show="!showExpandedSearch"
        :disabled="showExpandedSearch"
      ></v-text-field>
      <v-btn class="align-self-center mx-3" v-show="!showExpandedSearch" icon color="indigo">
        <v-icon @click="toggleExpandedSearch" v-if="!showExpandedSearch">mdi-chevron-down</v-icon>
        <v-icon @click="toggleExpandedSearch" v-else>mdi-chevron-up</v-icon>
      </v-btn>
    </v-container>
    <v-container v-show="showExpandedSearch" cols="12">
      <v-card>
        <v-card-title>
          Advance Query Search
          <v-btn-toggle v-model="mode" mandatory class="ml-auto">
            <v-btn value="and" small color="primary">AND</v-btn>
            <v-btn value="or" small color="primary">OR</v-btn>
          </v-btn-toggle>
        </v-card-title>
        <v-divider></v-divider>
        <v-card-text>
          <v-row>
            <v-col cols="12" class="d-flex">
              <v-select
                :items="availableQueries"
                item-text="name"
                item-value="name"
                v-model="selectedNewQuery"
                solo
                dense
                hide-details
                label="field"
              ></v-select>
              <v-btn color="success" @click="addQuery" :disabled="!selectedNewQuery">Add Query</v-btn>
            </v-col>
            <v-col cols="12" v-if="queries.length === 0">
              <v-alert text dense color="success" icon="mdi-information-outline" border="left"
                >You have no queries, please add at-least one!</v-alert
              >
            </v-col>
            <v-col v-else cols="12">
              <v-form ref="form" v-for="query in queries" v-bind:key="query.name">
                <query
                  :default="query.default"
                  :componentName="query.componentName"
                  :componentProps="query.props"
                  :name="query.name"
                  :defaultOperator="query.defaultOperator"
                  v-model="query.value"
                  :operators="query.operators"
                  @delete="deleteQuery"
                  @operatorChange="query.operator = $event"
                ></query>
              </v-form>
            </v-col>
          </v-row>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-btn color="info" text @click="showExpandedSearch = false" class="ml-auto">
            <v-icon>mdi-close</v-icon>
            Close
          </v-btn>
          <v-btn color="success" @click="advanceSearch" :disabled="queries.length === 0 || !valid" :loading="loading">
            <v-icon>mdi-magnify</v-icon>
            Search
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-container>
  </v-container>
</template>

<script>
import { POKEMON_LIST } from "../../lib/gpss/pokemon";
import { QUERIES_LIST } from "../../lib/gpss/queries";
import query from "./query";

export default {
  name: "gpss-search",
  components: { query },
  data: () => ({
    showExpandedSearch: false,
    queries: [],
    basicQuery: "",
    availableQueries: QUERIES_LIST,
    selectedNewQuery: "",
    mode: "and",
    valid: true,
  }),
  props: {
    loading: {
      type: Boolean,
      default: false,
    },
    gpssMode: {
      type: String,
      required: true,
    },
  },
  methods: {
    toggleExpandedSearch: function() {
      const vm = this;
      vm.showExpandedSearch = !vm.showExpandedSearch;
    },
    addQuery: function() {
      const vm = this;

      // get the selected query
      const query = vm.availableQueries.find((q) => q.name === vm.selectedNewQuery);
      // remove from available queries
      vm.availableQueries = vm.availableQueries.filter((q) => q.name !== query.name);
      // append to queries
      vm.queries.push(query);
      // clear the currently selected query
      vm.selectedNewQuery = "";
    },
    deleteQuery: function(name) {
      const vm = this;
      // Get from used queries first
      const query = vm.queries.find((q) => q.name === name);
      // Remove from queries
      vm.queries = vm.queries.filter((q) => q.name !== name);
      // Push query back to available queries
      vm.availableQueries.push(query);
    },
    basicSearch() {
      const vm = this;

      const value = vm.basicQuery;

      let query = {
        mode: "or",
        ot_name: value,
        nickname: value,
        ht_name: value,
        operators: [
          {
            operator: "IN",
            field: "ot_name",
          },
          {
            operator: "IN",
            field: "ht_name",
          },
          {
            operator: "IN",
            field: "nickname",
          },
        ],
      };
      if (value.length === 10 && Number.isInteger(parseInt(value))) {
        query = {
          mode: "or",
          download_code: value,
          operators: [
            {
              operator: "=",
              field: "download_code",
            },
          ],
        };

        if (vm.gpssMode === "bundle") {
          query = {
            mode: "or",
            download_codes: value,
            operators: [
              {
                operator: "IN",
                field: "download_codes",
              },
            ],
          };
        }
      }
      vm.$emit("search", query);
    },
    advanceSearch() {
      const vm = this;

      let queryJson = {
        mode: vm.mode,
      };

      let operators = [];

      vm.queries.forEach((query) => {
        if (!Array.isArray(query.value)) {
          queryJson[query.queryField] = query.value;
        } else {
          if (query.queryField == "min_max_level") {
            queryJson["min_level"] = query.value[0];
            queryJson["max_level"] = query.value[1];
          } else {
            let data = [];
            for (let value of query.value) {
              data.push(value);
            }
            queryJson[query.queryField] = data;
          }
        }

        if (query.queryField !== "min_max_level") {
          operators.push({
            operator: query.operator,
            field: query.queryField,
          });
        }
      });

      queryJson["operators"] = operators;
      vm.$emit("search", queryJson);
    },
  },
  computed: {
    pokemonList: function() {
      const vm = this;

      var list = POKEMON_LIST.filter((p) => p.available_generations.some((g) => vm.generations.indexOf(g) !== -1)).map(function(pokemon) {
        return pokemon.name;
      });

      return list;
    },
  },
  watch: {
    basicQuery() {
      if (this.basicQueryTimeout) clearTimeout(this.basicQueryTimeout);

      this.basicQueryTimeout = setTimeout(() => {
        this.basicSearch();
      }, 500);
    },
    queries: {
      deep: true,
      handler: function(a) {
        // queries can define a function that modifies their state dynamically (e.g
        // based off the selected values of another query). when the query data changes,
        // tell any queries that have custom update handlers about it, so they can update.
        let queries = a;
        if (this.$refs.form) {
          let valid = true;
          this.$refs.form.some((f) => {
            if (!f.validate()) {
              valid = false;
              return true;
            }
          });
          this.valid = valid;
        }
        for (let i in queries) {
          if (typeof queries[i].update === "function") {
            let q = queries[i].update(queries, queries[i]);
            if (JSON.stringify(q) === JSON.stringify(queries[i])) {
              continue;
            }
            this.queries = queries;
          }
        }

        if (a.length === 0) {
          this.$emit("search", {});
        }
      },
    },
  },
};
</script>

<style></style>

<template>
  <v-menu v-model="menu" :close-on-content-click="false" :nudge-width="200" offset-x>
    <template v-slot:activator="{ on, attrs }">
      <v-btn color="primary" dark v-bind="attrs" v-on="on">
        Sorting
      </v-btn>
    </template>

    <v-card>
      <v-list>
        <v-list-item>
          <v-list-item-content>
            <v-list-item-title>Sorting Method</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>

      <v-divider></v-divider>

      <v-list>
        <v-list-item>
          <v-list-item-action>
            <v-select :items="sortFields" v-model="currentSort" label="Sort By" hint="The field to sort by" solo dense />
          </v-list-item-action>
        </v-list-item>

        <v-list-item>
          <v-list-item-action>
            <v-switch v-model="direction" color="purple"></v-switch>
          </v-list-item-action>
          <v-list-item-title>{{ direction ? "ascending" : "descending" }}</v-list-item-title>
        </v-list-item>
      </v-list>

      <v-card-actions>
        <v-spacer></v-spacer>

        <v-btn text @click="cancelSort">
          Cancel
        </v-btn>
        <v-btn color="primary" text @click="sort">
          Sort
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-menu>
</template>

<script>
export default {
  name: "gpss-sort",

  data: () => ({
    menu: false,
    direction: false,
    currentSort: "latest",
    sortFields: [
      {
        text: "Latest",
        value: "latest",
      },
      {
        text: "Legality Status",
        value: "legality",
      },
      {
        text: "Popularity",
        value: "popularity",
      },
    ],
    originalSort: "latest",
    originalDirection: false,
  }),

  methods: {
    sort() {
      const vm = this;
      vm.menu = false;
      vm.originalSort = vm.currentSort;
      vm.originalDirection = vm.direction;

      const sortData = {
        sort_field: vm.currentSort,
        sort_direction: vm.direction,
      };

      vm.$emit("sort", sortData);
    },
    cancelSort() {
      const vm = this;
      vm.menu = false;
      vm.currentSort = vm.originalSort;
      vm.direction = vm.originalDirection;
    },
  },
};
</script>

<style></style>

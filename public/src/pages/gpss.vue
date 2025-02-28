<template>
  <div>
    <v-btn :color="$vuetify.theme.dark ? '' : 'pink'" fab bottom right fixed @click="uploading = true">
      <v-icon>mdi-plus</v-icon>
    </v-btn>
    <v-container>
      <h2>Welcome to GPSS</h2>
      <h4 class="subheading">We currently have {{ count.pokemon }} Pokemon and {{ count.bundles }} Bundles available for download!</h4>
      <template v-show="pages > 0">
        <v-switch
          @change="lockModeChange"
          :disabled="!modeChangeEnabled"
          v-model="mode"
          false-value="individual"
          true-value="bundle"
          :label="`GPSS Mode: ${mode}`"
        ></v-switch>
        <v-row no-gutters>
          <v-col cols="12">
            <Search @search="search" :loading="loading" :gpssMode="mode" />
          </v-col>
          <v-col cols="12" md="11" class="align-self-center">
            <v-pagination v-model="page" :page="page" :length="pages" :total-visible="5"></v-pagination>
          </v-col>
          <v-col cols="12" md="1">
            <Sort @sort="updateSort" />
          </v-col>
          <v-col cols="12" v-if="mode === 'individual'">
            <Pokemons
              @update-page-count="updatePageCount"
              :sort="sort"
              :searchQuery="searchQuery"
              :page="page"
              :resetPage="resetPage"
              @update-loading="updateLoading"
            />
          </v-col>
          <v-col cols="12" v-else>
            <Bundles
              @update-page-count="updatePageCount"
              :sort="sort"
              :searchQuery="searchQuery"
              :page="page"
              :resetPage="resetPage"
              @update-loading="updateLoading"
            />
          </v-col>
        </v-row>
      </template>
    </v-container>
    <Upload v-if="uploading" @cancel="closeUpload" />
  </div>
</template>

<script>
import { Pokemons, Bundles, Search, Sort, Upload } from "~/components/gpss";
export default {
  name: "GPSS",
  components: { Pokemons, Search, Sort, Upload, Bundles },
  data: () => ({
    pages: 0,
    page: 1,
    mode: "individual",
    modeChangeEnabled: true,
    uploading: false,
    showExpandedSearch: false,
    searchQuery: {},
    sort: { sort_field: "latest", sort_direction: false },
    loading: false,
    resetPage: false,
    count: {
      pokemon: 0,
      bundles: 0,
    },
  }),

  created() {
    const vm = this;

    vm.$api
      .get("/gpss/stats")
      .then((resp) => {
        vm.count = resp.data;
      })
      .catch((error) => {
        vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
      });
  },

  methods: {
    updatePageCount(pages) {
      const vm = this;
      vm.pages = pages;
      vm.resetPage = false;
    },
    toggleExpandedSearch: function() {
      const vm = this;
      vm.showExpandedSearch = !vm.showExpandedSearch;
    },
    search: function(query) {
      const vm = this;
      vm.resetPage = true;
      vm.page = 1;
      vm.searchQuery = query;
    },
    updateSort: function(sort) {
      const vm = this;
      vm.resetPage = true;
      vm.page = 1;
      vm.sort = sort;
    },
    updateLoading: function(loadingStatus) {
      this.loading = loadingStatus;
    },
    closeUpload({ reload }) {
      const vm = this;
      vm.uploading = false;
      if (reload) {
        vm.$router.go();
      }
    },
    lockModeChange() {
      const vm = this;
      vm.modeChangeEnabled = false;
      setTimeout(() => {
        vm.modeChangeEnabled = true;
      }, 1000);
    },
  },
};
</script>

<style scoped></style>

<template>
  <div>
    <v-simple-table fixed-header>
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">
              Upload Date
            </th>
            <th class="text-left">
              Download Code
            </th>
            <th class="text-left">
              Pokemons
            </th>
            <th class="text-left">
              Downloads
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(bundle, index) in bundles" :key="bundle.id">
            <td>{{ bundle.upload_date | date }}</td>
            <td>{{ bundle.download_code }}</td>
            <td v-if="$vuetify.breakpoint.mdAndUp">
              <p v-for="(code, index2) in bundle.download_codes" :key="index2">
                <v-icon :color="bundle.pokemons[index2].legality ? 'green' : 'red'">
                  mdi-{{ bundle.pokemons[index2].legality ? "check" : "close" }}
                </v-icon>
                <v-btn text @click="view(code)">{{ code }}</v-btn>
                <b>Gen: {{ bundle.pokemons[index2].generation }}</b>
              </p>
            </td>
            <td>{{ bundle.download_count }}</td>
            <td>
              <v-btn
                color="primary"
                v-if="$vuetify.breakpoint.smAndDown"
                :x-small="$vuetify.breakpoint.smAndDown"
                @click="showDialog(index)"
              >
                View PKMN
              </v-btn>
              <v-btn color="primary" :x-small="$vuetify.breakpoint.smAndDown" @click="download(index, $event)">Download</v-btn>
            </td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>
    <v-dialog v-if="$vuetify.breakpoint.smAndDown" v-model="dialog.show" width="500">
      <v-card>
        <v-card-title class="headline"> Viewing Bundle {{ dialog.bundle.download_code }} </v-card-title>

        <v-card-text>
          <v-list class="transparent">
            <v-list-item v-for="(download_code, index) in dialog.bundle.download_codes" :key="download_code">
              <v-list-item-title @click="view(download_code)">{{ download_code }}</v-list-item-title>

              <v-list-item-icon>
                <v-icon :color="dialog.bundle.pokemons[index].legality ? 'green' : 'red'">
                  mdi-{{ dialog.bundle.pokemons[index].legality ? "check" : "close" }}
                </v-icon>
              </v-list-item-icon>

              <v-list-item-subtitle class="text-right"> Generation: {{ dialog.bundle.pokemons[index].generation }} </v-list-item-subtitle>
            </v-list-item>
          </v-list>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" text @click="dialog.show = false">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { downloadAsZip } from "../../lib/gpss/download";

export default {
  name: "gpss-bundles",
  props: {
    page: {
      type: Number,
      default: 0,
    },
    searchQuery: {
      type: Object,
      default: {},
    },
    sort: {
      type: Object,
      required: true,
    },
    resetPage: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({
    pages: 0,
    bundles: [],
    loading: false,
    downloading: false,
    dialog: {
      show: false,
      bundle: {},
    },
  }),

  created() {
    this.getData();
  },
  methods: {
    showDialog(i) {
      const vm = this;
      vm.dialog.show = true;
      vm.dialog.bundle = vm.bundles[i];
    },
    download: function(index, event) {
      const vm = this;
      if (event.detail != 1) {
        return;
      }
      if (!vm.downloading) {
        vm.downloading = true;
        vm.$api
          .get(`/gpss/download/bundle/${vm.bundles[index].download_code}`)
          .then((resp) => {
            downloadAsZip(resp.data.pokemons, vm.bundles[index].download_code);
            vm.bundles[index].download_count++;
          })
          .catch((error) => {
            vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
          })
          .finally(() => {
            vm.downloading = false;
          });
      }
    },
    view: function(downloadCode) {
      this.$router.push(`/gpss/${downloadCode}`);
    },
    getData: function() {
      const vm = this;
      if (!vm.loading) {
        vm.loading = true;
        vm.$api
          .post(`/gpss/search/bundles?page=${vm.page}`, {
            ...vm.searchQuery,
            ...vm.sort,
          })
          .then((resp) => {
            vm.pages = resp.data.pages;
            vm.bundles = resp.data.bundles;
            vm.$emit("update-page-count", vm.pages);
            vm.loading = false;
          })
          .catch((error) => {
            vm.loading = false;
            vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
          });
      }
    },
  },
  watch: {
    page: function() {
      const vm = this;
      if (!vm.resetPage) {
        vm.getData();
      }
    },
    searchQuery: function() {
      const vm = this;
      vm.getData();
    },
    sort: function() {
      const vm = this;
      vm.getData();
    },
    loading: function() {
      this.$emit("update-loading", this.loading);
    },
  },
};
</script>

<style></style>

<template>
  <div>
    <v-subheader>Filtered Words</v-subheader>
    <div class="elevation-1" id="admin-repository">
      <v-toolbar flat dense>
        <v-toolbar-title class="grey--text">Manage Filtered Words</v-toolbar-title>
        <v-spacer></v-spacer>
        <v-dialog v-model="addWordDialog" max-width="300px">
          <template v-slot:activator="{ on, attrs }">
            <v-btn v-bind="attrs" v-on="on" icon><v-icon>mdi-plus</v-icon></v-btn>
          </template>

          <v-card>
            <v-card-title>Add Word to filter</v-card-title>
            <v-divider></v-divider>
            <v-text-field v-model="word" solo autofocus single-line required hide-details placeholder="Add Word to Filter"></v-text-field>
            <v-switch v-model="strict" label="Strict mode (auto reject)"></v-switch>
            <v-switch v-model="caseSensitive" label="Case Sensitive"></v-switch>
            <v-divider></v-divider>
            <v-card-actions>
              <v-btn color="blue darken-1" text @click="addWordDialog = false">Close</v-btn>
              <v-btn color="blue darken-1" text @click="addWord(word, strict, caseSensitive)">Save</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-toolbar>

      <v-divider></v-divider>
      <v-data-table
        :server-items-length="total"
        :options.sync="options"
        :loading="loading"
        dense
        :headers="headers"
        :items="words"
        :items-per-page="20"
        no-data-text="No words available"
      >
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon
            small
            class="mr-2"
            @click="
              deleteDialog = true;
              deletingWord = item.string;
              caseSensitive = item.case_sensitive;
            "
            >mdi-delete</v-icon
          >
        </template>
        <template v-slot:[`item.created_date`]="{ item }">
          <span>{{ item.created_date | date }}</span>
        </template>
        <template v-slot:[`item.modified_date`]="{ item }">
          <span>{{ item.modified_date | date }}</span>
        </template>
      </v-data-table>
    </div>
    <v-dialog v-model="deleteDialog" max-width="300px">
      <v-card>
        <v-card-title>Delete Word {{ deletingWord }}?</v-card-title>
        <v-card-actions>
          <v-btn color="blue darken-1" text @click="deleteDialog = false">Close</v-btn>
          <v-btn color="blue darken-1" text @click="deleteWord(deletingWord, caseSensitive)">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
export default {
  name: "admin-words",
  data: () => ({
    words: [],
    total: 0,
    loading: true,
    addWordDialog: false,
    deleteDialog: false,
    deletingWord: "",
    caseSensitive: false,
    word: "",
    strict: false,
    options: {
      page: 1,
      itemsPerPage: 10,
    },
    dialogData: {
      show: false,
      title: "",
      content: "",
    },
    headers: [
      { text: "Created", value: "created_date" },
      { text: "Word", value: "string" },
      { text: "Strict", value: "strict" },
      { text: "Case Sensitive", value: "case_sensitive" },
      { text: "Added By", value: "added_by" },
      { text: "Actions", value: "actions", sortable: false },
    ],
  }),
  watch: {
    options: {
      handler() {
        this.getWords();
      },
      deep: true,
    },
  },
  methods: {
    refreshData() {
      const vm = this;
      vm.getWords();
    },
    deleteWord: function(word, caseSensitive) {
      const formData = new FormData();
      formData.append("case_sensitive", caseSensitive);

      this.$api
        .delete(`/moderation/words/${word}`, { data: formData })
        .then((resp) => {
          this.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          this.refreshData();
        })
        .catch((error) => {
          this.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        })
        .finally(() => {
          this.deleteDialog = false;
          this.caseSensitive = false;
          this.deletingWord = "";
        });
    },
    addWord: function(word, strict, caseSensitive) {
      const formData = new FormData();

      formData.append("word", word);
      formData.append("strict", strict);
      formData.append("case_sensitive", caseSensitive);

      this.$api
        .post("/moderation/words", formData)
        .then((resp) => {
          this.$store.dispatch("showSnackbar", { text: resp.data.message, color: "green", timeout: 5000 });
          this.refreshData();
        })
        .catch((error) => {
          this.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        })
        .finally(() => {
          this.addWordDialog = false;
          this.word = "";
          this.strict = false;
          this.caseSensitive = false;
        });
    },
    getWords() {
      const vm = this;
      vm.loading = true;
      let sort = "";
      let desc = "";
      if (vm.options.sortBy[0]) {
        sort = `&sort=${vm.options.sortBy[0]}`;
        desc = `&sortDesc=${vm.options.sortDesc[0] ? "no" : "yes"}`;
      }
      let url = `/moderation/words?page=${vm.options.page}&amount=${vm.options.itemsPerPage}${sort}${desc}`;
      vm.$api
        .get(url)
        .then((resp) => {
          vm.words = resp.data.words ?? [];
          vm.pages = resp.data.pages;
          vm.total = resp.data.total;
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        })
        .finally(() => {
          this.loading = false;
        });
    },
  },
};
</script>

<style scoped>
.clickable {
  cursor: pointer;
}
</style>

<template>
  <v-container>
    <v-overlay :absolute="true" :value="project === null"></v-overlay>
    <div v-if="project !== null">
      <p class="text-h4">{{ project.name }}</p>
      <span class="text-caption mr-2">
        {{ project.stars }}
        <v-icon>mdi-star</v-icon>
      </span>
      <span class="text-caption mr-2">
        {{ project.commits }}
        <v-icon>mdi-source-commit</v-icon>
      </span>
      <span class="text-caption mr-2">
        {{ project.total_downloads.toLocaleString() }}
        <v-icon>mdi-download</v-icon>
      </span>
      <span class="text-caption">
        {{ project.forks }}
        <v-icon>mdi-source-fork</v-icon>
      </span>
      <p class="text-subtitle-1">{{ project.description }}</p>
      <v-btn color="success" @click="visitRepo">
        <v-icon left>
          mdi-github
        </v-icon>
        View On GitHub
      </v-btn>
      <v-btn color="primary" @click="dialog = true" v-if="project.latest_release_cia">
        <v-icon left>
          mdi-qrcode
        </v-icon>
        View CIA QR
      </v-btn>
      <v-divider />
      <div class="body-1 pb-3 project-readme" v-html="project.read_me">{{ project.description }}</div>
      <v-dialog v-model="dialog" width="unset" v-if="project.latest_release_cia">
        <v-card>
          <v-card-title class="headline"> {{ project.name }} CIA </v-card-title>
          <v-card-text>
            <vue-qrcode :value="project.latest_release_cia" :options="{ width: 300 }"></vue-qrcode>
          </v-card-text>
          <v-divider></v-divider>

          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn color="primary" text @click="dialog = false">
              Close
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </div>
  </v-container>
</template>

<script>
import VueQrcode from "@chenfengyuan/vue-qrcode";
export default {
  name: "project",
  data() {
    return {
      project: null,
      dialog: false,
    };
  },
  components: { VueQrcode },
  created() {
    this.getData();
  },
  methods: {
    getData: function() {
      this.$api
        .get(`/github/repo/${this.$route.params.project}`)
        .then((resp) => {
          this.project = resp.data.repo;
          document.title = this.project.name + " · FlagBrew";
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
          this.$router.push("/404").catch(() => {});
        });
    },
    visitRepo: function() {
      window.location.href = this.project.url;
    },
  },
  watch: {
    $route: function(to, from) {
      this.getData();
    },
  },
};
</script>

<style scoped>
.project-readme {
  padding-top: 1rem;
  overflow-x: auto;
}
.project-readme >>> img {
  max-width: 100%;
  height: auto;
}
</style>

<template>
  <v-container>
    <div class="text-h6">
      <v-chip :color="`orange ${$vuetify.theme.dark ? 'darken-3' : ''}`">{{ projects.length }}</v-chip> Here's our projects
      <v-row no-gutters>
        <v-col class="pa-2" v-for="project in projects" :key="project.id" cols="12" md="6" lg="4" xl="3" sm="6">
          <v-card class="mx-auto mb-2 equal-height">
            <v-list-item>
              <v-list-item-content>
                <v-list-item-title class="headline">{{ project.name }}</v-list-item-title>
              </v-list-item-content>
            </v-list-item>

            <v-card-text style="height: 80px;" height="100%">
              <div class="body-1">{{ project.description }}</div>
            </v-card-text>
            <v-divider light style="margin-top:15px;"></v-divider>
            <v-card-actions>
              <v-btn @click="viewProject(project.name)" text color="deep-purple accent-4">
                View Project
              </v-btn>
              <v-spacer></v-spacer>
              <div class="text-caption mr-2">
                {{ project.stars }}
                <v-icon small>mdi-star</v-icon>
              </div>
              <div class="text-caption mr-2">
                {{ project.commits }}
                <v-icon small>mdi-source-commit</v-icon>
              </div>
              <div class="text-caption mr-2">
                {{ project.total_downloads.toLocaleString() }}
                <v-icon small>mdi-download</v-icon>
              </div>
              <div class="text-caption">
                {{ project.forks }}
                <v-icon small>mdi-source-fork</v-icon>
              </div>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
    </div>
  </v-container>
</template>

<script>
export default {
  name: "github-repos",
  computed: {
    projects() {
      return this.$store.state.projects;
    },
  },
  methods: {
    viewProject(project) {
      this.$router.push("/projects/" + project);
    },
  },
};
</script>

<style scoped>
.equal-height {
  height: 210px;
  overflow-y: auto;
}

.bottom-right {
  right: 0;
}
</style>

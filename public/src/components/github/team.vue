<template>
  <v-container>
    <div class="text-h6">Meet the Team behind FlagBrew</div>
    <v-row no-gutters>
      <v-col class="pa-2" v-for="member in members" :key="member.id" cols="12" sm="4">
        <v-card class="mx-auto mb-2 equal-height">
          <v-list-item three-line>
            <v-list-item-content>
              <div class="font-weight-light mb-2 caption">GitHub ID # {{ member.github_id }}</div>
              <v-list-item-title v-if="member.name != ''" class="headline mb-1">{{ member.name }}</v-list-item-title>
              <v-list-item-title v-else class="headline mb-1">{{ member.username }}</v-list-item-title>
              <v-list-item-subtitle v-if="member.bio != ''">{{ member.bio }}</v-list-item-subtitle>
            </v-list-item-content>

            <v-list-item-avatar size=69 rounded>
              <v-img :src="member.avatar"></v-img>
            </v-list-item-avatar>
          </v-list-item>
          <v-card-actions class="card-bottom">
            <v-btn @click="visitProfile(member.url)" text>View on GitHub</v-btn>
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  name: "github-team",
    computed: {
    members() {
      return this.$store.state.members;
    },
  },
  methods: {
    visitProfile: function(url){
        window.location.href = url
    }
  },
};
</script>

<style>

</style>

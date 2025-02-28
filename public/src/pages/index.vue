<template>
  <div>
    <v-parallax :src="hero">
      <v-row align="center" justify="center">
        <v-col class="text-center" cols="12">
          <h1 class="display-1 font-weight-thin mb-4">
            FlagBrew
          </h1>
          <h4 class="subheading">
            Welcome to the home of PKSM, Checkpoint and GPSS!
          </h4>
        </v-col>
      </v-row>
    </v-parallax>
    <github-team></github-team>
    <v-divider></v-divider>
    <github-repos></github-repos>
  </div>
</template>

<script>
import GithubTeam from "~/components/github/team";
import GithubRepos from "~/components/github/repos";
import hero from "~/static/img/hero.png";
export default {
  name: "index",
  components: { GithubTeam, GithubRepos },
  data: () => ({
    hero: hero,
  }),
};
</script>

<style>
.equal-height {
  overflow-y: auto;
  height: 175px;
}

.card-bottom {
  position: absolute;
  bottom: 0;
}
</style>

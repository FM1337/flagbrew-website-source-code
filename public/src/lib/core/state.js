import Vue from "vue";
import Vuex from "vuex";
import http from "~/lib/core/http";
import constraints from "~/lib/core/constants";

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    auth: null,
    projects: [],
    members: [],
    check_auth: true,
    snackbar: {
      text: "",
      color: "",
      timeout: "",
    },
  },
  mutations: {
    set_auth(state, auth) {
      state.auth = auth;
    },
    set_repos(state, repos) {
      state.projects = repos.sort((a, b) => (a.name > b.name ? 1 : b.name > a.name ? -1 : 0));
    },
    set_members(state, members) {
      state.members = members;
    },
    set_snackbar(state, snackbar) {
      state.snackbar.text = snackbar.text;
      state.snackbar.color = snackbar.color;
      state.snackbar.timeout = snackbar.timeout;
    },
  },
  actions: {
    fetch_projects({ commit }) {
      return new Promise((resolve, reject) => {
        http
          .get("/github/repos")
          .then((resp) => {
            const repos = resp.data.repos.filter((repo) => {
              if (!constraints.BLACKLISTED_REPOS.includes(repo.name)) {
                return repo;
              }
            });
            commit("set_repos", repos);
            resolve(repos);
          })
          .catch((err) => {
            commit("set_repos", []);
            reject(err);
          });
      });
    },
    fetch_members({ commit }) {
      return new Promise((resolve, reject) => {
        http
          .get("/github/members")
          .then((resp) => {
            const members = resp.data.members;
            commit("set_members", members);
            resolve(members);
          })
          .catch((err) => {
            commit("set_members", []);
            reject(err);
          });
      });
    },
    get_auth({ commit, state }) {
      return new Promise((resolve, reject) => {
        http
          .get("/auth/self")
          .then((resp) => {
            const user = resp.data.authenticated ? resp.data.user : false;
            commit("set_auth", user);
            resolve(user);
          })
          .catch((err) => {
            commit("set_auth", false);
            // TODO: if error field, send notification?
            reject(err);
          });
      });
    },
    logout({ commit }) {
      new Promise((resolve, reject) => {
        http
          .get("/auth/logout")
          .then((resp) => {
            commit("set_auth", false);
            resolve(resp);
          })
          .catch((err) => {
            // TODO: if error field, send notification?
            reject(err);
          });
      });
    },
    showSnackbar({ commit }, snackbar) {
      commit("set_snackbar", snackbar);
    },
  },
  getters: {
    // isAuthed: state => !!state.auth,
    // auth: state => state.auth,
    projects: (state) => {
      return state.projects;
    },
    members: (state) => {
      return state.members;
    },
  },
});

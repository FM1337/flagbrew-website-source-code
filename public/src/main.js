import Vue from "vue";
import VueProgressBar from "vue-progressbar";
import VueVirtualScroller from "vue-virtual-scroller";
import "regenerator-runtime/runtime";
import VueGtag from "vue-gtag";

Vue.use(VueProgressBar, {
  color: "rgb(120, 96, 255)",
  failedColor: "red",
  thickness: "4px",
});

Vue.use(VueVirtualScroller);

import Vuetify from "vuetify";
import "vuetify/dist/vuetify.min.css";
import "@mdi/font/css/materialdesignicons.css";
Vue.use(Vuetify);

// disable ripples globally.
const VBtn = Vue.component("VBtn");
VBtn.options.props.ripple.default = false;
const VChip = Vue.component("VChip");
VChip.options.props.ripple.default = false;
const VCheckbox = Vue.component("VCheckbox");
VCheckbox.options.props.ripple.default = false;

import app from "~/app";
import router from "~/lib/core/router";
import state from "~/lib/core/state";

if (localStorage.getItem("ga-opt-out") !== "true") {
  Vue.use(
    VueGtag,
    {
      config: { id: "[REDACTED]" },
    },
    router
  );
}

Vue.config.productionTip = false;
// Vue.config.devtools = false

import http from "~/lib/core/http";
window.api = http;
Vue.prototype.$api = http;

import { checkIPBan, checkIPRestrict } from "~/lib/core/utils";
window.banCheck = checkIPBan;
Vue.prototype.$banCheck = checkIPBan;

window.restrictCheck = checkIPRestrict;
Vue.prototype.$restrictCheck = checkIPRestrict;

import { dateFilter } from "~/lib/core/filters";

export default new Vue({
  vuetify: new Vuetify({
    icons: { iconfont: "mdi" },
    theme: { dark: localStorage.getItem("dark") === "true" ? true : false },
  }),
  router,
  store: state,
  el: "#app",
  render: (h) => h(app),
  filters: {
    date: dateFilter,
  },
});

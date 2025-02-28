<template>
  <div>
    <v-data-table
      :headers="headers"
      item-key="name"
      :items="logs"
      class="elevation-1"
      :server-items-length="total"
      :options.sync="options"
      :footer-props="{
        'items-per-page-options': [1, 5, 10, 15, 20, 25, 50, 75, 100],
      }"
      :loading="loading"
      loading-text="Loading... Please wait"
    >
      <template v-slot:[`item.pokemon_data`]="{ item }">
        <v-btn x-small color="primary" @click="viewPokemonInfo(item.pokemon_data, item.download_code)">
          View Pokemon Info
        </v-btn>
      </template>
      <template v-slot:[`item.patron`]="{ item }">
        <v-btn v-if="item.patron" x-small color="primary" @click="viewPatronInfo(item.patron_code, item.patron_discord)">
          View Patron Info
        </v-btn>
        <span v-else>false</span>
      </template>
      <template v-slot:[`item.download_code`]="{ item }">
        <div
          :class="{
            'primary--text': !item.deleted && logType !== 'gpss_deletion' && !item.rejected,
            clickable: !item.deleted && logType !== 'gpss_deletion' && !item.rejected,
          }"
          @click="navigate(`/gpss/${item.download_code}`, !item.deleted && logType !== 'gpss_deletion' && !item.rejected)"
        >
          {{ item.download_code }}
        </div>
      </template>
      <template v-slot:[`item.actions`]="{ item }">
        <v-tooltip top v-if="logType === 'gpss_upload' && !item.deleted && item.approved">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small @click="$emit('delete', item)">
              mdi-delete
            </v-icon>
          </template>
          <span>Delete Pokemon</span>
        </v-tooltip>
        <v-tooltip top v-if="logType === 'gpss_upload' && !item.approved && !item.rejected">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small @click="$emit('reject', item.download_code)">
              mdi-cancel
            </v-icon>
          </template>
          <span>Reject Pokemon</span>
        </v-tooltip>
        <v-tooltip top v-if="logType === 'gpss_upload' && !item.approved && !item.rejected">
          <template v-slot:activator="{ on, attrs }">
            <v-icon v-bind="attrs" v-on="on" small @click="$emit('approve', item.download_code)">
              mdi-check
            </v-icon>
          </template>
          <span>Approve Pokemon</span>
        </v-tooltip>
        <v-tooltip
          top
          v-if="
            logType === 'gpss_upload' &&
              (!checkIP(item.uploader_ip ? item.uploader_ip : item.ip, 'ban') ||
                notBannedIPs.includes(item.uploader_ip ? item.uploader_ip : item.ip))
          "
        >
          <template v-slot:activator="{ on, attrs }">
            <v-icon small v-bind="attrs" v-on="on" @click="$emit('ban', item.uploader_ip ? item.uploader_ip : item.ip)">
              mdi-hammer
            </v-icon>
          </template>
          <span>Ban IP</span>
        </v-tooltip>
        <v-tooltip top v-if="bannedIPs.includes(item.uploader_ip ? item.uploader_ip : item.ip) || logType === 'banned'">
          <template v-slot:activator="{ on, attrs }">
            <v-icon small v-bind="attrs" v-on="on" @click="$emit('unban', item.uploader_ip ? item.uploader_ip : item.ip)">
              mdi-door-open
            </v-icon>
          </template>
          <span>Unban IP</span>
        </v-tooltip>
        <v-tooltip
          top
          v-if="
            logType === 'gpss_upload' &&
              (!checkIP(item.uploader_ip ? item.uploader_ip : item.ip, 'restrict') ||
                notRestrictedIPs.includes(item.uploader_ip ? item.uploader_ip : item.ip))
          "
        >
          <template v-slot:activator="{ on, attrs }">
            <v-icon small v-bind="attrs" v-on="on" @click="$emit('restrict', item.uploader_ip ? item.uploader_ip : item.ip)">
              mdi-lock
            </v-icon>
          </template>
          <span>Restrict Uploader</span>
        </v-tooltip>
        <v-tooltip
          top
          v-if="
            (logType === 'gpss_upload' && restrictedIPs.includes(item.uploader_ip ? item.uploader_ip : item.ip)) ||
              logType === 'restrictions'
          "
        >
          <template v-slot:activator="{ on, attrs }">
            <v-icon small v-bind="attrs" v-on="on" @click="$emit('unrestrict', item.uploader_ip ? item.uploader_ip : item.ip)">
              mdi-lock-open
            </v-icon>
          </template>
          <span>Unrestrict Uploader</span>
        </v-tooltip>
      </template>

      <template v-slot:[`item.date`]="{ item }">
        <span>{{ item.date | date(true) }}</span>
      </template>

      <template v-slot:[`item.original_expiry_date`]="{ item }">
        <span>{{ item.original_expiry_date | date(true) }}</span>
      </template>
    </v-data-table>

    <v-dialog v-model="dialogData.show" width="unset" @click:outside="clearDialog">
      <v-card>
        <v-card-title>
          <span class="headline">{{ dialogData.title }}</span>
        </v-card-title>
        <v-card-text>
          <pre style="overflow: auto;">{{ JSON.stringify(dialogData.content, null, "\t") }}</pre>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="clearDialog">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
export default {
  name: "admin-logs-logs",
  data: () => ({
    logs: [],
    bannedIPs: [],
    restrictedIPs: [],
    notRestrictedIPs: [],
    notBannedIPs: [],
    checkingBanIPs: [],
    checkingRestrictIPs: [],
    total: 0,
    loading: true,
    options: {
      page: 1,
      itemsPerPage: 10,
    },
    dialogData: {
      show: false,
      title: "",
      content: "",
    },
  }),
  props: {
    logType: {
      type: String,
      required: true,
    },
    headers: {
      type: Array,
      required: true,
    },
    reload: {
      type: Boolean,
      required: true,
    },
    pendingPokemonOnly: {
      type: Boolean,
      default: false,
    },
  },
  watch: {
    options: {
      handler() {
        this.getLogs();
      },
      deep: true,
    },
    reload: {
      handler(reload) {
        if (reload) {
          this.refreshData();
        }
      },
    },
  },
  methods: {
    refreshData() {
      const vm = this;
      vm.getLogs();
      vm.bannedIPs = [];
      vm.notBannedIPs = [];
      vm.restrictedIPs = [];
      vm.notRestrictedIPs = [];
      vm.$emit("refreshed");
    },
    checkIP: async function(ip, mode) {
      const vm = this;
      if (mode === "restrict") {
        if (!vm.restrictedIPs.includes(ip) && !vm.notRestrictedIPs.includes(ip)) {
          if (!vm.checkingRestrictIPs.includes(ip)) {
            vm.checkingRestrictIPs.push(ip);
            const restricted = await vm.$restrictCheck(ip);
            if (restricted) {
              vm.restrictedIPs.push(ip);
            } else {
              vm.notRestrictedIPs.push(ip);
            }
            vm.checkingRestrictIPs = vm.checkingRestrictIPs.filter((address) => address !== ip);
          }
          return vm.restrictedIPs.includes(ip);
        }
      } else {
        if (!vm.bannedIPs.includes(ip) && !vm.notBannedIPs.includes(ip)) {
          if (!vm.checkingBanIPs.includes(ip)) {
            vm.checkingBanIPs.push(ip);
            const banned = await vm.$banCheck(ip);
            if (banned) {
              vm.bannedIPs.push(ip);
            } else {
              vm.notBannedIPs.push(ip);
            }
            vm.checkingBanIPs = vm.checkingBanIPs.filter((address) => address !== ip);
          }
        }
        return vm.bannedIPs.includes(ip);
      }
    },
    getLogs() {
      const vm = this;
      vm.loading = true;
      let sort = "";
      let desc = "";
      if (vm.options.sortBy[0]) {
        sort = `&sort=${vm.options.sortBy[0]}`;
        desc = `&sortDesc=${vm.options.sortDesc[0] ? "no" : "yes"}`;
      }
      let url = `/moderation/logs?type=${vm.logType}&page=${vm.options.page}&amount=${vm.options.itemsPerPage}${sort}${desc}`;
      if (vm.logType === "gpss_upload" && vm.pendingPokemonOnly) {
        url += `&pending=yes`;
      }
      vm.$api
        .get(url)
        .then((resp) => {
          vm.logs = resp.data.logs ?? [];
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
    navigate(page, enabled) {
      const vm = this;
      if (vm.$router.history.current.fullPath !== page && enabled) {
        vm.$router.push(page);
      }
    },
    viewPokemonInfo(data, downloadCode) {
      const vm = this;
      vm.dialogData.title = `Viewing data for ${downloadCode}`;
      vm.dialogData.content = data;
      vm.dialogData.show = true;
    },
    viewPatronInfo(code, discord) {
      const vm = this;
      vm.dialogData.title = "Patreon Data";
      vm.dialogData.content = {
        patreonCode: code,
        discordID: discord,
      };
      vm.dialogData.show = true;
    },
    clearDialog() {
      const vm = this;
      vm.dialogData.show = false;
      vm.dialogData.title = "";
      vm.dialogData.content = "";
    },
  },
};
</script>

<style scoped>
.clickable {
  cursor: pointer;
}
</style>

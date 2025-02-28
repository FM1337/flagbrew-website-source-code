<template>
  <div>
    <v-row>
      <v-col cols="12" md="6">
        <v-card v-if="log">
          <v-card-title>Moderation Info</v-card-title>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Upload Date</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ log.date | date(true) }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Uploader IP</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ log.uploader_ip }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Upload Source</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{
                log.upload_source === "Discord" ? `Discord ${log.uploader_discord}` : log.upload_source
              }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Patron Upload</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ log.patron ? `Yes (${log.patron_code})` : "No" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Bundle Upload</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ log.bundle_upload ? `Yes (${log.bundle_code})` : "No" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>Upload Approved</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ log.approved ? "Yes" : "No" }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <v-list-item v-if="log.approved">
            <v-list-item-content>
              <v-list-item-title>Upload Approved By</v-list-item-title>
            </v-list-item-content>
            <v-list-item-content>
              <v-list-item-title>{{ log.approved_by }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card>
          <v-card-title>Moderation Actions</v-card-title>
          <v-list-item v-if="log && !log.approved">
            <v-list-item-content>
              <v-btn color="success" @click="approving = true">Approve Pokemon</v-btn>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>
              <v-btn color="error" @click="deleting = true" v-if="log && log.approved">Delete Pokemon</v-btn>
              <v-btn color="error" @click="rejecting = true" v-if="log && !log.approved">Reject Pokemon</v-btn>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content v-if="hasIP">
              <v-btn color="error" @click="banning = true" v-if="log && !isBanned">Ban Uploader</v-btn>
              <v-btn color="error" @click="unbanning = true" v-if="log && isBanned">Unban Uploader</v-btn>
            </v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content v-if="hasIP">
              <v-btn color="error" @click="restricting = true" v-if="log && !isRestricted">Restrict Uploader</v-btn>
              <v-btn color="error" @click="unrestricting = true" v-if="log && isRestricted">Unrestrict Uploader</v-btn>
            </v-list-item-content>
          </v-list-item>
        </v-card>
      </v-col>
    </v-row>
    <DeletePokemon
      :deletionData="log ? log : { download_code: downloadCode }"
      v-if="deleting"
      @close="deleting = false"
      @deleted="$router.push('/gpss')"
    />
    <BanIP :passedIP="log.uploader_ip" v-if="banning" @close="banning = false" @banned="banned" />
    <UnbanIP :passedIP="log.uploader_ip" v-if="unbanning" @close="unbanning = false" @unbanned="unbanned" />
    <ApprovePokemon :pokemonCode="downloadCode" v-if="approving" @close="approving = false" @approved="$router.go()" />
    <RejectPokemon :pokemonCode="downloadCode" v-if="rejecting" @close="rejecting = false" @rejected="$router.push('/gpss')" />
    <RestrictIP :passedIP="log.uploader_ip" v-if="restricting" @close="restricting = false" @restricted="restricted" />
    <UnrestrictIP :passedIP="log.uploader_ip" v-if="unrestricting" @close="unrestricting = false" @unrestricted="unrestricted" />
  </div>
</template>

<script>
import { ApprovePokemon, BanIP, DeletePokemon, RejectPokemon, RestrictIP, UnbanIP, UnrestrictIP } from "../../admin-actions/";

export default {
  name: "gpss-pokemon-moderation",
  data: () => ({
    log: null,
    deleting: false,
    banning: false,
    unbanning: false,
    isBanned: false,
    isRestricted: false,
    approving: false,
    rejecting: false,
    restricting: false,
    unrestricting: false,
    hasIP: false,
  }),
  props: {
    downloadCode: {
      type: String,
      required: true,
    },
  },
  components: { ApprovePokemon, BanIP, DeletePokemon, RejectPokemon, RestrictIP, UnbanIP, UnrestrictIP },
  created() {
    const vm = this;
    vm.$api
      .get(`/moderation/log/gpss_upload?query_field=download_code&query_value=${vm.downloadCode}`)
      .then((resp) => {
        vm.log = resp.data.log;
        if (vm.log.uploader_ip !== undefined) {
          vm.hasIP = true;
          vm.checkBanned();
          vm.checkRestricted();
        }
      })
      .catch((error) => {
        vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
      });
  },
  methods: {
    // TODO CLEAN THIS UP NEXT TIME
    restricted() {
      const vm = this;
      vm.isRestricted = true;
      vm.restricting = false;
    },
    unrestricted() {
      const vm = this;
      vm.isRestricted = false;
      vm.unrestricting = false;
    },
    unbanned() {
      const vm = this;
      vm.isBanned = false;
      vm.unbanning = false;
    },
    banned() {
      const vm = this;
      vm.isBanned = true;
      vm.banning = false;
    },
    checkBanned: async function() {
      const vm = this;
      vm.isBanned = await vm.$banCheck(vm.log.uploader_ip);
    },
    checkRestricted: async function() {
      const vm = this;
      vm.isRestricted = await vm.$restrictCheck(vm.log.uploader_ip);
    },
  },
};
</script>

<style></style>

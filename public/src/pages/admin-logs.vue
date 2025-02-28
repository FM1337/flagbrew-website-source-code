<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <!-- TODO ADD Filter Switch for showing pending approval pokemons only -->
        <h2>Upload Logs</h2>
        <v-switch label="Pending Pokemon Only" v-model="uploadPendingApprovalsOnly" @change="reloadLogs['upload'] = true"></v-switch>
        <Logs
          logType="gpss_upload"
          :reload="reloadLogs.upload"
          :headers="uploadHeaders"
          :pendingPokemonOnly="uploadPendingApprovalsOnly"
          @delete="openDialog($event, 'delete')"
          @ban="openDialog($event, 'ban')"
          @unban="openDialog($event, 'unban')"
          @approve="openDialog($event, 'approve')"
          @reject="openDialog($event, 'reject')"
          @refreshed="refreshed('upload')"
          @restrict="openDialog($event, 'restrict')"
          @unrestrict="openDialog($event, 'unrestrict')"
        />
      </v-col>
      <v-col cols="12">
        <h2>Failed Upload Logs</h2>
        <Logs logType="gpss_failed_upload" :reload="reloadLogs.failedUpload" :headers="failedUploadHeaders" />
      </v-col>
      <v-col cols="12">
        <h2>Deletion Logs</h2>
        <Logs logType="gpss_deletion" :reload="reloadLogs.deletion" :headers="deletionHeaders" @refreshed="refreshed('deletion')" />
      </v-col>
      <v-col cols="12">
        <h2>GPSS Clean Logs</h2>
        <Logs logType="gpss_clean" :reload="reloadLogs.gpssClean" :headers="gpssCleanHeaders" @refreshed="refreshed('gpssClean')" />
      </v-col>
      <v-col cols="12">
        <h2>Banned IP Logs</h2>
        <v-btn color="primary" dark class="mb-2" @click="openDialog('', 'ban')">
          Ban IP
        </v-btn>
        <Logs
          logType="banned"
          @ban="openDialog($event, 'ban')"
          @unban="openDialog($event, 'unban')"
          :reload="reloadLogs.ban"
          :headers="bannedHeaders"
          @refreshed="refreshed('ban')"
        />
      </v-col>
      <v-col cols="12">
        <h2>Unban Logs</h2>
        <Logs logType="unban" :reload="reloadLogs.unban" :headers="unbanHeaders" @refreshed="refreshed('unban')" />
      </v-col>
      <v-col cols="12">
        <h2>Setting Change Logs</h2>
        <Logs
          logType="setting_change"
          :reload="reloadLogs.settingChange"
          :headers="settingChangeHeaders"
          @refreshed="refreshed('settingChanged')"
        />
      </v-col>
      <v-col cols="12">
        <h2>Restricted IPs</h2>
        <v-btn color="primary" dark class="mb-2" @click="openDialog('', 'restrict')">
          Restrict IP
        </v-btn>
        <Logs
          logType="restrictions"
          @unrestrict="openDialog($event, 'unrestrict')"
          :reload="reloadLogs.restricted"
          :headers="restrictedHeaders"
          @refreshed="refreshed('restricted')"
        />
      </v-col>
      <v-col cols="12">
        <h2>Unrestricted IP Logs</h2>
        <Logs
          logType="unrestrict"
          :reload="reloadLogs.unrestricted"
          :headers="unrestrictLogHeaders"
          @refreshed="refreshed('unrestricted')"
        />
      </v-col>
      <v-col cols="12">
        <h2>Patreon Build Delete Logs</h2>
        <Logs logType="build_delete" :reload="reloadLogs.patreonBuildDelete" :headers="patreonBuildDeleteHeaders" />
      </v-col>
      <v-col cols="12">
        <h2>Word Filter Delete Logs</h2>
        <Logs logType="word_delete" :reload="reloadLogs.wordFilterDelete" :headers="wordFilterDeleteHeaders" />
      </v-col>
    </v-row>
    <DeletePokemon
      v-if="dialogs.delete"
      :deletingPokemon="dialogData"
      :deletionData="dialogData"
      @close="closeDialog('delete')"
      @deleted="closeDialog('delete', true, 'deletion', 'upload')"
    />
    <BanIP v-if="dialogs.ban" @close="closeDialog('ban')" @banned="closeDialog('ban', true, 'upload', 'ban')" :passedIP="dialogData" />
    <UnbanIP
      v-if="dialogs.unban"
      @close="closeDialog('unban')"
      @unbanned="closeDialog('unban', true, 'upload', 'unban', 'ban')"
      :passedIP="dialogData"
    />
    <ApprovePokemon
      v-if="dialogs.approve"
      @close="closeDialog('approve')"
      @approved="closeDialog('approve', true, 'upload')"
      :pokemonCode="dialogData"
    />
    <RejectPokemon
      v-if="dialogs.reject"
      @close="closeDialog('reject')"
      @rejected="closeDialog('reject', true, 'upload')"
      :pokemonCode="dialogData"
    />
    <RestrictIP
      v-if="dialogs.restrict"
      @close="closeDialog('restrict')"
      @restricted="closeDialog('restrict', true, 'upload', 'restricted')"
      :passedIP="dialogData"
    />
    <UnrestrictIP
      v-if="dialogs.unrestrict"
      @close="closeDialog('unrestrict')"
      @unrestricted="closeDialog('unrestrict', true, 'upload', 'restricted', 'unrestricted')"
      :passedIP="dialogData"
    />
  </v-container>
</template>

<script>
import { Logs } from "../components/admin-logs/";
import { ApprovePokemon, BanIP, DeletePokemon, RejectPokemon, RestrictIP, UnbanIP, UnrestrictIP } from "../components/admin-actions/";
export default {
  name: "admin-logs",
  data: () => ({
    dialogs: {
      delete: false,
      approve: false,
      ban: false,
      unban: false,
      reject: false,
      restrict: false,
      unrestrict: false,
    },
    dialogData: null,
    reloadLogs: {
      upload: false,
      failedUpload: false,
      deletion: false,
      ban: false,
      unban: false,
      settingChange: false,
      gpssClean: false,
      patreonBuildDelete: false,
      wordFilterDelete: false,
      restricted: false,
      unrestricted: false,
    },
    uploadPendingApprovalsOnly: false,
    defaultHeaders: [
      {
        text: "Log Date",
        align: "start",
        value: "date",
      },
    ],
    uploadHeaders: [
      {
        text: "Uploader IP",
        value: "uploader_ip",
      },
      {
        text: "Upload Source",
        value: "upload_source",
      },
      {
        text: "Discord User",
        value: "uploader_discord",
      },
      {
        text: "Deleted",
        value: "deleted",
      },
      {
        text: "Pokemon Data",
        value: "pokemon_data",
      },
      {
        text: "Approved",
        value: "approved",
      },
      {
        text: "Approved By",
        value: "approved_by",
      },
      {
        text: "Rejected",
        value: "rejected",
      },
      {
        text: "Rejected By",
        value: "rejected_by",
      },
      {
        text: "Rejected Reason",
        value: "rejected_reason",
      },
      {
        text: "Patron",
        value: "patron",
      },
      {
        text: "Bundle Upload",
        value: "bundle_upload",
      },
      {
        text: "Download Code",
        value: "download_code",
      },
      {
        text: "Bundle Code",
        value: "bundle_code",
      },
      { text: "Actions", value: "actions", sortable: false },
    ],
    failedUploadHeaders: [
      {
        text: "Uploader IP",
        value: "uploader_ip",
      },
      {
        text: "Upload Source",
        value: "upload_source",
      },
      {
        text: "Discord Username",
        value: "uploader_discord",
      },
      {
        text: "Rejection/Failure Reason",
        value: "rejected_reason",
      },
      {
        text: "Patron Upload",
        value: "patron",
      },
    ],
    deletionHeaders: [
      {
        text: "Deleted By",
        value: "deleted_by",
      },
      {
        text: "Deletion Reason",
        value: "deletion_reason",
      },
      {
        text: "Entity Type",
        value: "entity_type",
      },
      {
        text: "Download Code",
        value: "download_code",
      },
    ],
    unbanHeaders: [
      {
        text: "IP",
        value: "ban.ip",
      },
      {
        text: "Unbanned By",
        value: "unbanned_by",
      },
      {
        text: "Original Ban Date",
        value: "ban.date",
      },
      {
        text: "Original Ban Reason",
        value: "ban.ban_reason",
      },
      {
        text: "Original Banned By",
        value: "ban.banned_by",
      },
    ],
    bannedHeaders: [
      {
        text: "IP",
        value: "ip",
      },
      {
        text: "Ban Reason",
        value: "ban_reason",
      },
      {
        text: "Banned By",
        value: "banned_by",
      },
      { text: "Actions", value: "actions", sortable: false },
    ],
    settingChangeHeaders: [
      {
        text: "Setting Name",
        value: "setting",
      },
      {
        text: "Original Value",
        value: "original_value",
      },
      {
        text: "New Value",
        value: "new_value",
      },
      {
        text: "Modified By",
        value: "modified_by",
      },
    ],
    gpssCleanHeaders: [
      {
        text: "Amount Deleted",
        value: "deleted",
      },
      {
        text: "Amount Reset",
        value: "reset",
      },
      {
        text: "Amount Failed",
        value: "failed",
      },
    ],
    patreonBuildDeleteHeaders: [
      {
        text: "Commit Hash",
        value: "commit_hash",
      },
      {
        text: "Filename",
        value: "filename",
      },
      {
        text: "Original Expiry Date",
        value: "original_expiry_date",
      },
    ],
    unrestrictLogHeaders: [
      {
        text: "Unrestricted By",
        value: "unrestricted_by",
      },
      {
        text: "Original Restriction",
        value: "original_restriction",
      },
    ],

    wordFilterDeleteHeaders: [
      {
        text: "Deleted By",
        value: "deleted_by",
      },
      {
        text: "Original Word",
        value: "original_word",
      },
    ],

    restrictedHeaders: [
      {
        text: "IP",
        value: "ip",
      },
      {
        text: "Restricted By",
        value: "restricted_by",
      },
      {
        text: "Restricted Reason",
        value: "restricted_reason",
      },
      { text: "Actions", value: "actions", sortable: false },
    ],
  }),
  created() {
    this.failedUploadHeaders = [...this.defaultHeaders, ...this.failedUploadHeaders];
    this.uploadHeaders = [...this.defaultHeaders, ...this.uploadHeaders];
    this.deletionHeaders = [...this.defaultHeaders, ...this.deletionHeaders];
    this.unbanHeaders = [...this.defaultHeaders, ...this.unbanHeaders];
    this.bannedHeaders = [...this.defaultHeaders, ...this.bannedHeaders];
    this.settingChangeHeaders = [...this.defaultHeaders, ...this.settingChangeHeaders];
    this.gpssCleanHeaders = [...this.defaultHeaders, ...this.gpssCleanHeaders];
    this.patreonBuildDeleteHeaders = [...this.defaultHeaders, ...this.patreonBuildDeleteHeaders];
    this.unrestrictLogHeaders = [...this.defaultHeaders, ...this.unrestrictLogHeaders];
    this.wordFilterDeleteHeaders = [...this.defaultHeaders, ...this.wordFilterDeleteHeaders];
  },
  components: { Logs, ApprovePokemon, BanIP, DeletePokemon, RejectPokemon, RestrictIP, UnbanIP, UnrestrictIP },
  methods: {
    openDialog(data, dialog) {
      this.dialogData = data;
      this.dialogs[dialog] = true;
    },
    closeDialog(dialog, reload = false, ...logs) {
      const vm = this;
      vm.dialogs[dialog] = false;
      vm.dialogData = null;
      if (reload) {
        logs.forEach((log) => {
          vm.reloadLogs[log] = true;
        });
      }
    },
    refreshed(log) {
      this.reloadLogs[log] = false;
    },
  },
};
</script>

<style></style>

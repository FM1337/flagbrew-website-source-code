<template>
  <div>
    <v-subheader>Settings</v-subheader>
    <div class="elevation-1" id="admin-repository">
      <v-toolbar flat dense>
        <v-toolbar-title class="grey--text">Manage Settings</v-toolbar-title>
        <v-spacer></v-spacer>
      </v-toolbar>

      <v-divider></v-divider>
      <v-data-table :loading="loading" dense :headers="headers" :items="settings" :items-per-page="20" no-data-text="No settings available">
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon small class="mr-2" @click="$emit('edit', item)">mdi-pencil</v-icon>
        </template>
        <template v-slot:[`item.created_date`]="{ item }">
          <span>{{ item.created_date | date }}</span>
        </template>
        <template v-slot:[`item.modified_date`]="{ item }">
          <span>{{ item.modified_date | date }}</span>
        </template>
      </v-data-table>
    </div>
  </div>
</template>
<script>
export default {
  name: "admin-site-settings",
  data: () => ({
    loading: true,
    editSettingDialog: false,
    headers: [
      { text: "Name", value: "name" },
      { text: "Description", value: "description" },
      { text: "Value Type", value: "type" },
      { text: "Current Value", value: "value" },
      { text: "System Variable", value: "system_variable" },
      { text: "Created By", value: "created_by" },
      { text: "Created Date", value: "created_date" },
      { text: "Modified Date", value: "modified_date" },
      { text: "Actions", value: "actions", sortable: false },
    ],
    settings: [],
  }),
  props: {
    reload: {
      type: Boolean,
      required: true,
    },
  },
  watch: {
    reload: {
      handler(reload) {
        if (reload) {
          this.refreshData();
        }
      },
    },
  },
  created() {
    this.initialize();
  },
  methods: {
    initialize: function() {
      const vm = this;
      this.$api
        .get("/moderation/settings")
        .then((resp) => {
          vm.settings = [];
          Object.keys(resp.data.settings).forEach((key) => vm.settings.push(resp.data.settings[key]));
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        })
        .finally(() => {
          this.loading = false;
        });
    },
    refreshData() {
      this.initialize();
      this.$emit("reloaded");
    },
  },
};
</script>

<style></style>

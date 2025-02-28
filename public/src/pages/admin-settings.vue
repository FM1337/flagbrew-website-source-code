<template>
  <v-container>
    <v-row justify-center>
      <v-col cols="12">
        <Users />
        <Settings @edit="editSetting" @reloaded="reloadSettings = false" :reload="reloadSettings" />
        <Words />
        <EditSetting v-if="showSettingEdit" @updatedSetting="updateSetting" @close="closeSetting" :setting="setting" />
        <Gift />
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
// example: https://codesandbox.io/s/o29j95wx9
import { Users, Settings, Gift, Words } from "../components/admin-settings";
import { EditSetting } from "../components/admin-actions";
export default {
  name: "admin-settings",
  components: { Users, Settings, Gift, EditSetting, Words },
  data: () => ({
    setting: null,
    showSettingEdit: false,
    reloaded: false,
    reloadSettings: false,
  }),
  created() {
    console.log(`Site was built on ${process.env.BUILD_DATE} `);
  },
  methods: {
    editSetting(setting) {
      this.setting = setting;
      this.showSettingEdit = true;
    },
    updateSetting() {
      this.reloadSettings = true;
      this.closeSetting();
    },
    closeSetting() {
      this.showSettingEdit = false;
      this.setting = null;
    },
  },
};
</script>

<style scoped></style>

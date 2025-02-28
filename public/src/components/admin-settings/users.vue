<template>
  <div>
    <v-subheader>Users</v-subheader>
    <div class="elevation-1" id="admin-repository">
      <v-toolbar flat dense>
        <v-toolbar-title class="grey--text">Manage Users</v-toolbar-title>
        <v-spacer></v-spacer>

        <v-tooltip top>
          <template v-slot:activator="{ on, attrs }">
            <v-btn icon v-bind="attrs" v-on="on" href="https://caius.github.io/github_id/" target="_blank"
              ><v-icon color="grey lighten-1">mdi-github</v-icon></v-btn
            >
          </template>
          <span>Find GitHub ID by username</span>
        </v-tooltip>
        <v-dialog v-model="addUserDialog" max-width="300px">
          <template v-slot:activator="{ on, attrs }">
            <v-btn v-bind="attrs" v-on="on" icon><v-icon>mdi-plus</v-icon></v-btn>
          </template>

          <v-card>
            <v-card-title>Add GitHub ID</v-card-title>
            <v-divider></v-divider>
            <v-text-field v-model="addUser" solo autofocus single-line hide-details placeholder="Add user by GitHub ID"></v-text-field>
            <v-divider></v-divider>
            <v-card-actions>
              <v-btn color="blue darken-1" text @click="addUserDialog = false">Close</v-btn>
              <v-btn color="blue darken-1" text @click="createUser(addUser)">Save</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-toolbar>

      <v-divider></v-divider>
      <v-data-table :loading="loading" dense :headers="headers" :items="users" :items-per-page="20" no-data-text="No users configured">
        <template v-slot:[`item.avatar_url`]="{ item }">
          <v-avatar v-if="item.avatar_url" size="25"><img :src="item.avatar_url"/></v-avatar>
        </template>
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon small class="mr-2" @click="deleteUser(item)">mdi-delete</v-icon>
        </template>
      </v-data-table>
    </div>
  </div>
</template>

<script>
export default {
  name: "admin-users",
  data: () => ({
    loading: true,
    addUserDialog: false,
    addUser: "",
    headers: [
      { text: "Avatar", value: "avatar_url" },
      { text: "ID", value: "id" },
      { text: "Github ID", value: "github_id" },
      { text: "Name", value: "name" },
      { text: "Username", value: "username" },
      { text: "Actions", value: "actions", sortable: false },
    ],
    users: [],
  }),
  created: function() {
    this.initialize();
  },
  methods: {
    initialize: function() {
      this.$api
        .get("/users")
        .then((resp) => {
          this.users = resp.data.users;
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        })
        .finally(() => {
          this.loading = false;
        });
    },
    deleteUser: function(user) {
      this.$api
        .delete(`/users/${user.id}`)
        .then((resp) => {
          this.initialize();
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        });
    },
    createUser: function(user) {
      const data = {
        github_id: parseInt(user, 10),
      };
      this.$api
        .post("/users", data)
        .then((resp) => {
          this.initialize();
        })
        .catch((error) => {
          vm.$store.dispatch("showSnackbar", { text: error.response.data.error, color: "red", timeout: 5000 });
        })
        .finally(() => {
          this.addUserDialog = false;
        });
    },
  },
};
</script>

<style scoped></style>

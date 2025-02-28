<template>
  <v-row justify="center">
    <v-dialog v-model="dialog" fullscreen hide-overlay transition="dialog-bottom-transition">
      <v-card>
        <v-toolbar dark color="primary">
          <v-btn icon dark @click="cancel">
            <v-icon>mdi-close</v-icon>
          </v-btn>
          <v-toolbar-title>Uploading to GPSS</v-toolbar-title>
        </v-toolbar>
        <v-stepper v-model="step">
          <v-stepper-header>
            <v-stepper-step :complete="step > 1" step="1">
              Upload Type
            </v-stepper-step>
            <v-divider></v-divider>
            <v-stepper-step :complete="step > 2" step="2">
              Upload Information
            </v-stepper-step>
            <v-divider></v-divider>
            <v-stepper-step step="3" :complete="step > 3">
              Upload
            </v-stepper-step>
            <v-divider></v-divider>
            <v-stepper-step step="4" :complete="step === 4 && !uploading && successfulUpload">
              Finished
            </v-stepper-step>
          </v-stepper-header>
          <!-- Step 1, select the upload type -->
          <v-stepper-items>
            <v-stepper-content step="1">
              <v-card-text>
                <p class="display-1 text--primary">Choose Upload Type</p>
                <div class="text--primary">
                  Bundle upload allows you to upload up to 6 Pokemon at once <b>but requires at-least 2 pokemon</b><br />
                  To upload a single Pokemon please choose individual
                </div>
                <v-radio-group v-model="uploadType" row>
                  <v-radio label="Individual" value="individual"></v-radio>
                  <v-radio label="Bundle" value="bundle"></v-radio>
                </v-radio-group>
              </v-card-text>
              <v-btn color="primary" @click="navigate(2)" :disabled="!canNavNext">
                Continue
              </v-btn>
              <v-btn text @click="cancel(false)">
                Cancel
              </v-btn>
            </v-stepper-content>
            <!-- end step 1 -->
            <!-- Step 2 -->

            <v-stepper-content step="2">
              <v-row v-if="uploadType === 'individual'">
                <v-col cols="12" md="4">
                  <v-select
                    :items="items"
                    v-model="uploadInfo.generation"
                    label="Generation"
                    @change="uploadInfo.pokemon = null"
                  ></v-select>
                </v-col>
                <v-col cols="12" md="4">
                  <v-file-input
                    :accept="getAcceptableFileExtension(uploadInfo.generation)"
                    label="Pokemon"
                    v-model="uploadInfo.pokemon"
                  ></v-file-input>
                </v-col>
                <v-col cols="12" md="4">
                  <v-text-field v-model="uploadInfo.patreonCode" color="teal">
                    <template v-slot:label>
                      <div>Patreon Code <small>(optional)</small></div>
                    </template>
                  </v-text-field>
                </v-col>
              </v-row>
              <v-row v-else>
                <v-col cols="12" md="6">
                  <v-text-field v-model.number="uploadInfo.count" type="number" label="Pokemon Count" max="6" min="2"></v-text-field>
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field v-model="uploadInfo.patreonCode" color="teal">
                    <template v-slot:label>
                      <div>Patreon Code <small>(optional)</small></div>
                    </template>
                  </v-text-field>
                </v-col>
                <v-col cols="12" v-if="uploadInfo.count >= 2 && uploadInfo.count <= 6">
                  <v-row v-for="i in uploadInfo.count" :key="i">
                    <v-col cols="12" md="6">
                      <v-select
                        :items="items"
                        v-model="uploadInfo.generations[i - 1]"
                        label="Generation"
                        @change="uploadInfo.pokemons[i - 1] = null"
                      ></v-select>
                    </v-col>
                    <v-col cols="12" md="6">
                      <v-file-input
                        :accept="getAcceptableFileExtension(uploadInfo.generations[i - 1])"
                        label="Pokemon"
                        v-model="uploadInfo.pokemons[i - 1]"
                      ></v-file-input>
                    </v-col>
                  </v-row>
                </v-col>
              </v-row>
              <v-btn color="primary" @click="navigate(1)">
                Back
              </v-btn>
              <v-btn color="primary" @click="navigate(3)" :disabled="!canNavNext">
                Continue
              </v-btn>
              <v-btn text @click="cancel(false)">
                Cancel
              </v-btn>
            </v-stepper-content>

            <!-- end step 2 -->
            <!-- step 3 -->
            <v-stepper-content step="3">
              <v-card-text v-if="step === 3">
                <p class="display-1 text--primary">Please Verify Information</p>
                <div class="text--primary">
                  If the following info looks correct, feel free to hit <b>upload</b><br />
                  Otherwise, hit back and adjust what you need to adjust.
                </div>
                <div v-if="uploadType === 'individual'">
                  <v-list-item three-line>
                    <v-list-item-content>
                      <v-list-item-title>{{ uploadInfo.pokemon.name }}</v-list-item-title>
                      <v-list-item-title> Generation: {{ uploadInfo.generation }} </v-list-item-title>
                      <v-list-item-title> Patreon Code: {{ uploadInfo.patreonCode ? uploadInfo.patreonCode : "None" }} </v-list-item-title>
                    </v-list-item-content>
                  </v-list-item>
                </div>

                <div v-else>
                  <v-list-item-title> Patreon Code: {{ uploadInfo.patreonCode ? uploadInfo.patreonCode : "None" }} </v-list-item-title>
                  <v-row>
                    <v-col cols="12" md="3" v-for="i in uploadInfo.count" :key="i">
                      <v-list-item three-line>
                        <v-list-item-content>
                          <v-list-item-title>{{ uploadInfo.pokemons[i - 1].name }}</v-list-item-title>
                          <v-list-item-title> Generation: {{ uploadInfo.generations[i - 1] }} </v-list-item-title>
                        </v-list-item-content>
                      </v-list-item>
                    </v-col>
                  </v-row>
                </div>
              </v-card-text>
              <v-btn color="primary" @click="navigate(2)">
                Back
              </v-btn>
              <v-btn color="primary" @click="navigate(4)">
                Upload
              </v-btn>
              <v-btn text @click="cancel(false)">
                Cancel
              </v-btn>
            </v-stepper-content>
            <!-- end step 3 -->
            <!-- step 4 -->
            <v-stepper-content step="4">
              <v-card-text v-if="uploading">
                <p class="display-1 text--primary"><v-progress-circular indeterminate color="primary"></v-progress-circular> Uploading</p>
                <div class="text--primary">
                  Please wait...
                </div>
              </v-card-text>
              <v-card-text v-else>
                <p v-if="successfulUpload" class="display-1 text--primary">
                  <v-icon color="green" large>mdi-check</v-icon> Upload Sucessful!
                </p>
                <p v-else class="display-1 text--primary"><v-icon color="red" large>mdi-close</v-icon> Upload Failed!</p>
                <div class="text--primary" v-if="successfulUpload && approved">
                  Your Pokemon {{ uploadType === "bundle" ? "have" : "has" }} been uploaded, if you'd like to upload more then please click
                  the <b>Upload Again</b> button, otherwise <br />
                  {{ uploadType === "bundle" ? "" : "click the view pokemon button to view your pokemon or " }}
                  click the close button to exit this dialog and reload GPSS.
                </div>
                <div class="text--primary" v-else-if="successfulUpload && !approved">
                  Your Pokemon {{ uploadType === "bundle" ? "have" : "has" }} been uploaded however, they are currently being held for
                  manual review, if you'd like to upload more then please click the <b>Upload Again</b> button, otherwise click the close
                  button to exit this dialog and reload GPSS.
                </div>
                <div class="text--primary" v-else>
                  An error while uploading was encountered, information about this error can be found below: <br />
                  <b>{{ result.error }}</b> <br />
                  You can either try again later, or reach out to us on Discord or at
                  <a href="mailto:[REDACTED]">[REDACTED]</a>
                </div>
              </v-card-text>
              <v-btn
                color="primary"
                v-if="successfulUpload && approved && uploadType == 'individual'"
                @click="$router.push(`/gpss/${result.code}`)"
              >
                View Pokemon
              </v-btn>
              <v-btn color="primary" v-if="successfulUpload" @click="restart">
                Upload Again
              </v-btn>
              <v-btn text @click="cancel(true)">
                Close
              </v-btn>
            </v-stepper-content>
            <!-- end step 4 -->
          </v-stepper-items>
        </v-stepper>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
export default {
  name: "gpss-upload",
  data() {
    return {
      dialog: true,
      uploadType: "",
      oldUploadType: "",
      step: 1,
      uploadInfo: {
        generation: "",
        pokemon: null,
        patreonCode: "",
      },
      items: ["1", "2", "3", "4", "5", "6", "7", "LGPE", "8"],
      uploading: false,
      successfulUpload: false,
      result: {},
    };
  },
  methods: {
    restart() {
      const vm = this;
      vm.step = 1;
      vm.uploadType = "";
      vm.uploadInfo = {
        generation: "",
        pokemon: null,
        patreonCode: "",
      };
      vm.successfulUpload = false;
      vm.approved = false;
      vm.uploadType = "";
      vm.oldUploadType = "";
    },
    getAcceptableFileExtension(generation) {
      switch (generation) {
        case "1":
          return ".pk1";
        case "2":
          return ".pk2";
        case "3":
          return ".pk3";
        case "4":
          return ".pk4";
        case "5":
          return ".pk5";
        case "6":
          return ".pk6";
        case "7":
          return ".pk7";
        case "LGPE":
          return ".pb7";
        case "8":
          return ".pk8";
        default:
          return ".pk1,.pk2,.pk3,.pk4,.pk5,.pk6,.pk7,.pk8,.pb7";
      }
    },
    cancel(reload = false) {
      const vm = this;
      vm.dialog = false;
      setTimeout(() => {
        vm.$emit("cancel", { reload: reload });
      }, 250);
    },
    navigate(stage) {
      const vm = this;
      if (stage === 2 && vm.oldUploadType !== vm.uploadType) {
        // set the old upload type
        vm.oldUploadType = vm.uploadType;
        // Check the mode and set upload info struct
        if (vm.uploadType === "individual") {
          vm.uploadInfo = {
            generation: "",
            pokemon: null,
            patreonCode: "",
          };
        } else {
          vm.uploadInfo = {
            count: 2,
            generations: [],
            pokemons: [],
            patreonCode: "",
          };
        }
      } else if (stage === 4) {
        vm.uploading = true;
        // call upload
        vm.upload();
      }
      vm.step = stage;
    },
    upload() {
      const vm = this;
      // Create formData
      const formData = new FormData();
      let headers = {
        "Content-Type": "multipart/form-data",
        source: `Web Browser: ${navigator.userAgent}`,
      };

      if (vm.uploadInfo.patreonCode !== "") {
        headers["patreon"] = vm.uploadInfo.patreonCode;
      }

      if (vm.uploadType === "bundle") {
        vm.uploadInfo.pokemons.forEach((pk, i) => {
          formData.append(`pkmn${i + 1}`, pk);
        });
        headers["generations"] = vm.uploadInfo.generations;
        headers["count"] = vm.uploadInfo.count;
      } else {
        formData.append("pkmn", vm.uploadInfo.pokemon);
        headers["generation"] = vm.uploadInfo.generation;
      }

      // Now post the data
      vm.$api
        .post(`/gpss/upload/${vm.uploadType === "bundle" ? "bundle" : "pokemon"}`, formData, {
          headers: headers,
        })
        .then((resp) => {
          vm.result = {
            code: resp.data.code,
          };
          vm.successfulUpload = true;
          vm.approved = resp.data.approved;
        })
        .catch((error) => {
          vm.result = {
            error: error.response.data.error,
          };
        })
        .finally(() => {
          vm.uploading = false;
        });
    },
  },
  watch: {
    dialog: function() {
      const vm = this;
      setTimeout(() => {
        vm.$emit("cancel", { reload: vm.successfulUpload });
      }, 250);
    },
  },
  computed: {
    canNavNext: function() {
      const vm = this;
      if (vm.step === 1) {
        return vm.uploadType !== "";
      } else if (vm.step === 2) {
        if (vm.uploadType === "individual") {
          return vm.uploadInfo.generation !== "" && vm.uploadInfo.pokemon !== null;
        } else {
          for (let i = 0; i < vm.uploadInfo.count; i++) {
            if (
              vm.uploadInfo.pokemons[i] === null ||
              vm.uploadInfo.pokemons[i] === undefined ||
              vm.uploadInfo.generations[i] === "" ||
              vm.uploadInfo.generations[i] === null ||
              vm.uploadInfo.generations[i] === undefined
            ) {
              return false;
            }
          }
          return true;
        }
      }

      return true;
    },
  },
};
</script>

<style></style>

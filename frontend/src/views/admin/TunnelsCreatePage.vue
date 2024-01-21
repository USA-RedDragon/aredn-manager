<template>
  <div>
    <PVToast />
    <form @submit.prevent="handleSubmit(!v$.$invalid)">
      <Card>
        <template #title>Create Tunnel</template>
        <template #content>
          <h3 class="card-section-header">Tunnel Type</h3>
          <br/>
          <div class="flex align-items-center card-section">
            <RadioButton v-model="wireguard" inputId="wireguard1" name="wireguard" :value="false" />
              <label for="wireguard1" class="ml-2">VTun</label>
          </div>
          <div class="flex align-items-center card-section">
            <RadioButton v-model="wireguard" inputId="wireguard2" name="wireguard" :value="true" />
            <label for="wireguard2" class="ml-2">Wireguard</label>
          </div>
          <br/>
          <h3 class="card-section-header">Connection Type</h3>
          <br/>
          <div class="flex align-items-center card-section">
            <RadioButton v-model="tunnelType" name="tunnelType" value="server" />
            <label class="ml-2">Server - Provide a tunnel to another node</label>
          </div>
          <div class="flex align-items-center card-section">
            <RadioButton v-model="tunnelType" name="tunnelType" value="client" />
            <label class="ml-2">Client - Connect to another node's tunnel</label>
          </div>
          <br />
          <h3 class="card-section-header">Connection Settings</h3>
          <br/>
          <div class="card-section" v-if="tunnelType == 'server'">
            <span class="p-float-label">
              <InputText
                id="hostname"
                type="text"
                v-model="v$.hostname.$model"
                :class="{
                  'p-invalid': v$.hostname.$invalid && submitted,
                }"
              />
              <label
                for="hostname"
                :class="{ 'p-error': v$.hostname.$invalid && submitted }"
                >Hostname</label
              >
            </span>
            <span v-if="v$.hostname.$error && submitted">
              <span v-for="(error, index) of v$.hostname.$errors" :key="index">
                <small class="p-error">{{ error.$message }}</small>
                <br />
              </span>
            </span>
            <span v-else>
              <small
                v-if="
                  (v$.hostname.$invalid && submitted) ||
                  v$.hostname.$pending.$response
                "
                class="p-error"
                >{{ v$.hostname.required.$message }}
                <br />
              </small>
            </span>
            <br />
          </div>
          <div class="card-section" v-if="tunnelType == 'server' && !wireguard">
            <PVCheckbox
              id="generatepassword"
              :binary="true"
              v-model="this.generatePassword"
              @change="handleGeneratePassword()"
            />&nbsp;
            <label for="generatepassword">Generate Password</label>
            <br v-if="!generatePassword" />
            <br v-if="!generatePassword" />
          </div>
          <div class="card-section" v-if="tunnelType == 'client'">
            <span class="p-float-label">
              <InputText
                id="server"
                type="text"
                v-model="v$.server.$model"
                :class="{
                  'p-invalid': v$.server.$invalid && submitted,
                }"
              />
              <label
                for="server"
                :class="{ 'p-error': v$.server.$invalid && submitted }"
                >Server Address</label
              >
            </span>
            <span v-if="v$.server.$error && submitted">
              <span v-for="(error, index) of v$.server.$errors" :key="index">
                <small class="p-error">{{ error.$message }}</small>
                <br />
              </span>
            </span>
            <span v-else>
              <small
                v-if="
                  (v$.server.$invalid && submitted) ||
                  v$.server.$pending.$response
                "
                class="p-error"
                >{{ v$.server.required.$message }}
                <br />
              </small>
            </span>
            <br />
            <span class="p-float-label">
              <InputText
                id="network"
                type="text"
                v-model="v$.network.$model"
                :class="{
                  'p-invalid': v$.network.$invalid && submitted,
                }"
              />
              <label
                for="network"
                :class="{ 'p-error': v$.network.$invalid && submitted }"
                >Network</label
              >
            </span>
            <span v-if="v$.network.$error && submitted">
              <span v-for="(error, index) of v$.network.$errors" :key="index">
                <small class="p-error">{{ error.$message }}</small>
                <br />
              </span>
            </span>
            <span v-else>
              <small
                v-if="
                  (v$.network.$invalid && submitted) ||
                  v$.network.$pending.$response
                "
                class="p-error"
                >{{ v$.network.required.$message }}
                <br />
              </small>
            </span>
            <br />
          </div>
          <div class="card-section">
            <span class="p-float-label" v-if="(!wireguard && !generatePassword) || tunnelType=='client'">
              <InputText
                id="password"
                type="password"
                v-model="v$.password.$model"
                :class="{
                  'p-invalid': v$.password.$invalid && submitted,
                }"
              />
              <label
                v-if="!wireguard"
                for="password"
                :class="{ 'p-error': v$.password.$invalid && submitted }"
                >Password</label
              >
              <label
                v-else
                for="password"
                :class="{ 'p-error': v$.password.$invalid && submitted }"
                >Key</label
              >
            </span>
            <span v-if="v$.password.$error && submitted">
              <span v-for="(error, index) of v$.password.$errors" :key="index">
                <small class="p-error">{{ error.$message }}</small>
                <br />
              </span>
            </span>
            <span v-else>
              <small
                v-if="
                  !generatePassword &&
                  ((v$.password.$invalid && submitted) ||
                  v$.password.$pending.$response)
                "
                class="p-error"
                >{{ v$.password.required.$message }}
                <br />
              </small>
            <br />
            </span>
          </div>
          <div class="card-section" v-if="tunnelType == 'server' && !wireguard">
            <span class="p-float-label" v-if="!generatePassword">
              <InputText
                id="confirmPassword"
                type="password"
                v-model="v$.confirmPassword.$model"
                :class="{
                  'p-invalid': v$.confirmPassword.$invalid && submitted,
                }"
              />
              <label
                for="confirmPassword"
                :class="{ 'p-error': v$.confirmPassword.$invalid && submitted }"
                >Confirm Password</label
              >
            </span>
            <span v-if="!generatePassword && v$.confirmPassword.$error && submitted">
              <span
                v-for="(error, index) of v$.confirmPassword.$errors"
                :key="index"
              >
                <small class="p-error">{{ error.$message }}</small>
                <br />
              </span>
            </span>
            <span v-else>
              <small
                v-if="
                  !generatePassword &&
                  ((v$.confirmPassword.$invalid && submitted) ||
                  v$.confirmPassword.$pending.$response)
                "
                class="p-error"
                >{{ v$.confirmPassword.required.$message }}
                <br />
              </small>
            </span>
          </div>
        </template>
        <template #footer>
          <div class="card-footer">
            <PVButton
              class="p-button-raised p-button-rounded"
              icon="pi pi-user"
              type="submit"
              label="Submit"
            />
          </div>
        </template>
      </Card>
    </form>
  </div>
</template>

<script>
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import Checkbox from 'primevue/checkbox';
import Card from 'primevue/card';
import RadioButton from 'primevue/radiobutton';
import API from '@/services/API';

import { useVuelidate } from '@vuelidate/core';
import { required, sameAs, requiredIf, minLength, maxLength, ipAddress } from '@vuelidate/validators';

export default {
  components: {
    InputText,
    PVButton: Button,
    PVCheckbox: Checkbox,
    RadioButton,
    Card,
  },
  setup: () => ({ v$: useVuelidate() }),
  created() {
    this.handleGeneratePassword();
  },
  mounted() {},
  data: function() {
    return {
      wireguard: false,
      hostname: '',
      server: '',
      network: '',
      password: '',
      confirmPassword: '',
      generatePassword: true,
      tunnelType: 'server',
      submitted: false,
    };
  },
  validations() {
    return {
      hostname: {
        required: requiredIf(this.tunnelType == 'server'),
        minLength: minLength(3),
        maxLength: maxLength(63),
      },
      password: {
        required,
        minLength: minLength(3),
      },
      confirmPassword: {
        required: requiredIf(this.tunnelType == 'server'),
        sameAs: sameAs(this.password),
      },
      server: {
        required: requiredIf(this.tunnelType == 'client'),
        minLength: minLength(3),
      },
      network: {
        required: requiredIf(this.tunnelType == 'client'),
        ipAddress: ipAddress,
      },
    };
  },
  methods: {
    handleGeneratePassword() {
      if (!this.generatePassword) {
        this.password = '';
        this.confirmPassword = '';
      } else {
        // Create a 6 character random password matching [a-zA-Z]
        this.password = this.generateRandomPassword();
        this.confirmPassword = this.password;
      }
    },
    generateRandomPassword() {
      // This is A-Z and a-z, just shuffled to avoid any patterns in the random number generator
      const characters = 'fjyaYxnQBEzplkLiZuPhvXAVNOKSMFrdgGJTqRbDwIHWUCectosm';
      let password = '';
      for (let i = 0; i < 6; i++) {
        password += characters.charAt(Math.floor(Math.random() * characters.length));
      }
      return password;
    },
    async generatePrivateKey() {
      try {
        const res = await API.get('/wireguard/genkey');
        return res.data.key;
      } catch (err) {
        console.error(err);
      }
    },
    handleSubmit(isFormValid) {
      this.submitted = true;
      if (!isFormValid && this.v$.$errors.length > 0) {
        return;
      }

      if (this.tunnelType == 'client') {
        // parse server address as a hostname and optional port
        const serverParts = this.server.split(':');
        if (serverParts.length > 2) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Server Address must be in the format hostname:port',
            life: 3000,
          });
          return;
        }

        if (serverParts[0].length > 253) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Server Address must be less than 254 characters',
            life: 3000,
          });
          return;
        }

        if (!/^[A-Za-z0-9-\\.]+$/.test(serverParts[0])) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Server Address hostname must be alphanumeric, \'.\', or \'-\'',
            life: 3000,
          });
          return;
        }

        if (serverParts.length == 2) {
          if (serverParts[1] < 1 || serverParts[1] > 65535) {
            this.$toast.add({
              severity: 'error',
              summary: 'Error',
              detail: 'Server Address port must be between 1 and 65535',
              life: 3000,
            });
            return;
          }
        }

        API.post('/tunnels', {
          hostname: this.server.trim(),
          password: this.password.trim(),
          ip: this.network.trim(),
          client: true,
        })
          .then((res) => {
            this.$toast.add({
              severity: 'success',
              summary: 'Success',
              detail: res.data.message,
              life: 3000,
            });
            this.$router.push('/admin/tunnels');
          })
          .catch((err) => {
            console.error(err);
            if (err.response && err.response.data && err.response.data.error) {
              this.$toast.add({
                severity: 'error',
                summary: 'Error',
                detail: err.response.data.error,
                life: 3000,
              });
            } else {
              this.$toast.add({
                severity: 'error',
                summary: 'Error',
                detail: 'An unknown error occurred',
                life: 3000,
              });
            }
          });
      } else if (this.tunnelType == 'server') {
        if (this.confirmPassword != this.password) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: `Passwords do not match`,
            life: 3000,
          });
          return;
        }

        if (this.hostname.length < 3) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Hostname must be at least 3 characters',
            life: 3000,
          });
          return;
        }

        if (this.hostname.length > 63) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Hostname must be less than 64 characters',
            life: 3000,
          });
          return;
        }

        if (!/^[A-Za-z0-9-]+$/.test(this.hostname)) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Hostname must be alphanumeric or -',
            life: 3000,
          });
          return;
        }

        API.post('/tunnels', {
          hostname: this.hostname.trim(),
          password: this.password.trim(),
          client: false,
        })
          .then((res) => {
            this.$toast.add({
              severity: 'success',
              summary: 'Success',
              detail: res.data.message,
              life: 3000,
            });
            this.$router.push('/admin/tunnels');
          })
          .catch((err) => {
            console.error(err);
            if (err.response && err.response.data && err.response.data.error) {
              this.$toast.add({
                severity: 'error',
                summary: 'Error',
                detail: err.response.data.error,
                life: 3000,
              });
            } else {
              this.$toast.add({
                severity: 'error',
                summary: 'Error',
                detail: 'An unknown error occurred',
                life: 3000,
              });
            }
          });
      }
    },
  },
};
</script>

<style scoped></style>

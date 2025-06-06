<template>
  <div>
    <PVToast />
    <form @submit.prevent="handleSubmit(!v$.$invalid)">
      <Card>
        <template #title>Create Wireguard Tunnel</template>
        <template #content>
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
            <span class="p-float-label" v-if="tunnelType=='client'">
              <InputText
                id="password"
                type="password"
                v-model="v$.password.$model"
                :class="{
                  'p-invalid': v$.password.$invalid && submitted,
                }"
              />
              <label
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
import Card from 'primevue/card';
import RadioButton from 'primevue/radiobutton';
import API from '@/services/API';

import { useVuelidate } from '@vuelidate/core';
import { required, requiredIf, minLength, maxLength } from '@vuelidate/validators';

export default {
  components: {
    InputText,
    PVButton: Button,
    RadioButton,
    Card,
  },
  setup: () => ({ v$: useVuelidate() }),
  created() {
  },
  mounted() {},
  data: function() {
    return {
      hostname: '',
      server: '',
      network: '',
      password: '',
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
        required: required,
        minLength: minLength(44*3),
      },
      server: {
        required: requiredIf(this.tunnelType == 'client'),
        minLength: minLength(3),
      },
      network: {
        required: requiredIf(this.tunnelType == 'client'),
      },
    };
  },
  methods: {
    handleSubmit(isFormValid) {
      this.submitted = true;
      if (!isFormValid && this.v$.$errors.length > 0) {
        return;
      }

      if (this.tunnelType == 'client') {
        // parse server address as a hostname and optional port
        const networkParts = this.network.split(':');
        if (networkParts.length !== 2) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Network must be in the format hostname:port',
            life: 3000,
          });
          return;
        }

        if (networkParts[0].length > 253) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Network must be less than 254 characters',
            life: 3000,
          });
          return;
        }

        if (!/^[A-Za-z0-9-\\.]+$/.test(networkParts[0])) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Network hostname must be alphanumeric, \'.\', or \'-\'',
            life: 3000,
          });
          return;
        }

        if (networkParts.length == 2) {
          if (networkParts[1] < 1 || networkParts[1] > 65535) {
            this.$toast.add({
              severity: 'error',
              summary: 'Error',
              detail: 'Network port must be between 1 and 65535',
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
          wireguard: true,
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

        if (this.hostname.length < 3) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Hostname must be at least 3 characters',
            life: 3000,
          });
          return;
        }

        API.post('/tunnels', {
          hostname: this.hostname.trim(),
          client: false,
          wireguard: true,
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

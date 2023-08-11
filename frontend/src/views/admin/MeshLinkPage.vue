<template>
  <div>
    <PVToast />
    <form @submit.prevent="handleSubmit(!v$.$invalid)">
      <Card>
        <template #title>Create Supernode Connection</template>
        <template #content>
          <span class="p-float-label">
            <InputText
              id="name"
              type="text"
              v-model="v$.name.$model"
              :class="{
                'p-invalid': v$.name.$invalid && submitted,
              }"
              style="width: 85%;"
            />
            .mesh
            <label
              for="name"
              :class="{ 'p-error': v$.name.$invalid && submitted }"
              >DNS Zone</label
            >
          </span>
          <span v-if="v$.name.$error && submitted">
            <span v-for="(error, index) of v$.name.$errors" :key="index">
              <small class="p-error">{{ error.$message }}</small>
              <br />
            </span>
          </span>
          <span v-else>
            <small
              v-if="
                (v$.name.$invalid && submitted) ||
                v$.name.$pending.$response
              "
              class="p-error"
              >{{ v$.name.required.$message }}
              <br />
            </small>
          </span>
          <br />
          <span class="p-float-label">
            <TextArea
              id="ipsBox"
              autoResize rows="5"
              v-model="v$.ipsBox.$model"
              :class="{
                'p-invalid': v$.ipsBox.$invalid && submitted,
              }"
              lines="5"
            />
            <label
              for="name"
              :class="{ 'p-error': v$.name.$invalid && submitted }"
              >IP addresses, one per line</label
            >
          </span>
          <span v-if="v$.ipsBox.$error && submitted">
            <span v-for="(error, index) of v$.ipsBox.$errors" :key="index">
              <small class="p-error">{{ error.$message }}</small>
              <br />
            </span>
          </span>
          <span v-else>
            <small
              v-if="
                (v$.ipsBox.$invalid && submitted) ||
                v$.ipsBox.$pending.$response
              "
              class="p-error"
              >{{ v$.ipsBox.required.$message }}
              <br />
            </small>
          </span>
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
import TextArea from 'primevue/textarea';
import Button from 'primevue/button';
import Card from 'primevue/card';
import API from '@/services/API';

import { useVuelidate } from '@vuelidate/core';
import { required, ipAddress } from '@vuelidate/validators';

export default {
  components: {
    InputText,
    PVButton: Button,
    TextArea,
    Card,
  },
  setup: () => ({ v$: useVuelidate() }),
  created() {},
  mounted() {},
  data: function() {
    return {
      name: '',
      ipsBox: '',
      ips: [],
      submitted: false,
    };
  },
  validations() {
    return {
      name: {
        required,
      },
      ipsBox: {
        required,
      },
    };
  },
  methods: {
    handleSubmit(isFormValid) {
      this.submitted = true;
      if (!isFormValid) {
        return;
      }

      // Validate ipsBox has valid IPs
      this.ips = this.ipsBox.split('\n');
      for (let i = 0; i < this.ips.length; i++) {
        const ip = this.ips[i].trim();
        this.ips[i] = ip;
        if (!ipAddress.$validator(ip, null, { required: true })) {
          this.$toast.add({
            severity: 'error',
            summary: 'Error',
            detail: 'Invalid IP address: ' + ip,
            life: 3000,
          });
          return;
        }
      }

      // and that the dns name isn't bad
      if (this.name.length < 3) {
        this.$toast.add({
          severity: 'error',
          summary: 'Error',
          detail: 'DNS name must be at least 3 characters',
          life: 3000,
        });
        return;
      }

      if (this.name.length > 63) {
        this.$toast.add({
          severity: 'error',
          summary: 'Error',
          detail: 'DNS name must be less than 64 characters',
          life: 3000,
        });
        return;
      }

      if (!/^[A-Za-z0-9-]+$/.test(this.name)) {
        this.$toast.add({
          severity: 'error',
          summary: 'Error',
          detail: 'DNS name must be alphanumeric or -',
          life: 3000,
        });
        return;
      }

      API.post('/meshes', {
        name: this.name.trim(),
        ips: this.ips,
      })
        .then((res) => {
          this.$toast.add({
            severity: 'success',
            summary: 'Success',
            detail: res.data.message,
            life: 3000,
          });
          this.$router.push('/admin/meshes');
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
    },
  },
};
</script>

<style scoped></style>

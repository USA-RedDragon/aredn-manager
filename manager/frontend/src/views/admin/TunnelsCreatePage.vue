<template>
  <div>
    <PVToast />
    <form @submit.prevent="handleSubmit(!v$.$invalid)">
      <Card>
        <template #title>Create Tunnel</template>
        <template #content>
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
          <PVCheckbox
            id="generatepassword"
            :binary="true"
            v-model="this.generatePassword"
            @change="handleGeneratePassword()"
          />&nbsp;
          <label for="generatepassword">Generate Password</label>
          <br v-if="!generatePassword" />
          <br v-if="!generatePassword" />
          <span class="p-float-label" v-if="!generatePassword">
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
              >Password</label
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
                (v$.password.$invalid && submitted) ||
                v$.password.$pending.$response
              "
              class="p-error"
              >{{ v$.password.required.$message }}
              <br />
            </small>
          <br />
          </span>
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
          <span v-if="v$.confirmPassword.$error && submitted">
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
                (v$.confirmPassword.$invalid && submitted) ||
                v$.confirmPassword.$pending.$response
              "
              class="p-error"
              >{{ v$.confirmPassword.required.$message }}
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
import Button from 'primevue/button';
import Checkbox from 'primevue/checkbox';
import Card from 'primevue/card';
import API from '@/services/API';

import { useVuelidate } from '@vuelidate/core';
import { required, sameAs } from '@vuelidate/validators';

export default {
  components: {
    InputText,
    PVButton: Button,
    PVCheckbox: Checkbox,
    Card,
  },
  setup: () => ({ v$: useVuelidate() }),
  created() {
    this.handleGeneratePassword();
  },
  mounted() {},
  data: function() {
    return {
      hostname: '',
      password: '',
      confirmPassword: '',
      generatePassword: true,
      submitted: false,
    };
  },
  validations() {
    return {
      hostname: {
        required,
      },
      password: {
        required,
      },
      confirmPassword: {
        required,
        sameAs: sameAs(this.password),
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
    handleSubmit(isFormValid) {
      this.submitted = true;
      if (!isFormValid) {
        return;
      }

      if (this.confirmPassword != this.password) {
        this.$toast.add({
          severity: 'error',
          summary: 'Error',
          detail: `Passwords do not match`,
          life: 3000,
        });
        return;
      }
      API.post('/tunnels', {
        hostname: this.hostname.trim().toUpperCase(),
        password: this.password.trim(),
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
    },
  },
};
</script>

<style scoped></style>

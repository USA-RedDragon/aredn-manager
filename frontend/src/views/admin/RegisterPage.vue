<template>
  <div>
    <PVToast />
    <form @submit.prevent="handleRegister(!v$.$invalid)">
      <Card>
        <template #title>Register</template>
        <template #content>
          <span class="p-float-label">
            <InputText
              id="username"
              type="text"
              v-model="v$.username.$model"
              :class="{
                'p-invalid': v$.username.$invalid && submitted,
              }"
            />
            <label
              for="username"
              :class="{ 'p-error': v$.username.$invalid && submitted }"
              >Username</label
            >
          </span>
          <span v-if="v$.username.$error && submitted">
            <span v-for="(error, index) of v$.username.$errors" :key="index">
              <small class="p-error">{{ error.$message }}</small>
              <br />
            </span>
          </span>
          <span v-else>
            <small
              v-if="
                (v$.username.$invalid && submitted) ||
                v$.username.$pending.$response
              "
              class="p-error"
              >{{ v$.username.required.$message }}
              <br />
            </small>
          </span>
          <br />
          <span class="p-float-label">
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
          </span>
          <br />
          <span class="p-float-label">
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
              label="Register"
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
import API from '@/services/API';

import { useVuelidate } from '@vuelidate/core';
import { required, sameAs } from '@vuelidate/validators';

export default {
  components: {
    InputText,
    PVButton: Button,
    Card,
  },
  setup: () => ({ v$: useVuelidate() }),
  created() {},
  mounted() {},
  data: function() {
    return {
      username: '',
      password: '',
      confirmPassword: '',
      submitted: false,
    };
  },
  validations() {
    return {
      username: {
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
    handleRegister(isFormValid) {
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
      API.post('/users', {
        username: this.username.trim(),
        password: this.password.trim(),
      })
        .then((res) => {
          this.$toast.add({
            severity: 'success',
            summary: 'Success',
            detail: res.data.message,
            life: 3000,
          });
          this.$router.push('/admin/users');
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

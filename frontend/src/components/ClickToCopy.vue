<template>
  <span>
    <PVButton
      @click="copyToClipboard"
      :label=text
      :class="{
        'click-to-copy': true,
        'click-to-copy--copied': this.copied,
      }"
      text
    />
  </span>
</template>

<script>
import Button from 'primevue/button';

export default {
  props: {
    text: {
      type: String,
      default: 'Click to copy',
    },
    copy: {
      type: String,
      default: '',
    },
  },
  components: {
    PVButton: Button,
  },
  data: function() {
    return {
      copied: false,
    };
  },
  mounted() {
  },
  unmounted() {
  },
  methods: {
    copyToClipboard() {
      if ('navigator' in window && 'clipboard' in window.navigator) {
        navigator.clipboard.writeText(this.copy).then(() => {
          this.copied = true;
          setTimeout(() => {
            this.copied = false;
          }, 1000);
        });
      } else {
        const el = document.createElement('textarea');
        el.value = this.copy;
        document.body.appendChild(el);
        el.select();
        document.execCommand('copy');
        document.body.removeChild(el);
        this.copied = true;
        setTimeout(() => {
          this.copied = false;
        }, 1000);
      }
    },
  },
  computed: {
  },
};
</script>

<style scoped>
.click-to-copy {
  cursor: pointer;
  user-select: none;
}

.click-to-copy--copied {
  color: green;
}
</style>

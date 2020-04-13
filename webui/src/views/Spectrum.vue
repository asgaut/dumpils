<template>
  <div>
    <div style="text-align: center;">
      <label for="source">Input:</label>
      <select id="source" v-model="source">
        <option value="loc">LOC</option>
        <option value="gp">GP</option>
      </select>&nbsp;
      <label for="stage">Stage:</label>
      <select id="stage" v-model="stage">
        <option value="if">IF</option>
        <option value="lf">LF</option>
      </select>
    </div>
    <br />
    <SvgSpectrum :url="`http://localhost:3344/spectrum?source=${source}&amp;stage=${stage}`" />
  </div>
</template>

<script>
// @ is an alias to /src
import SvgSpectrum from "@/components/SvgSpectrum.vue";

export default {
  name: "Spectrum",
  data: function() {
    return {
      source: "loc",
      stage: "if"
    };
  },
  components: {
    SvgSpectrum
  }
};
</script>

<style>
</style>
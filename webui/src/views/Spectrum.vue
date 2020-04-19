<template>
  <div>
    <div style="text-align: center;">
      <br />
      <label for="source">Input:</label>
      <select id="source" v-model="source">
        <option value="loc">LOC</option>
        <option value="gp">GP</option>
      </select>&nbsp;
      <label for="stage">Stage:</label>
      <select id="stage" v-model="stage">
        <option v-for="(s, k) in stages" :key="k" :value="k">{{s.name}}</option>
      </select>
    </div>
    <br />
    <SvgSpectrum
      :fs="stages[stage].fs"
      :url="`http://localhost:3344/spectrum?source=${source}&amp;stage=${stage}`"
    />
  </div>
</template>

<script>
// @ is an alias to /src
import SvgSpectrum from "@/components/SvgSpectrum.vue";

const stages = {
  if: { name: "IF", fs: 1310720 },
  lf: { name: "LF", fs: 81920 }
};

export default {
  name: "Spectrum",
  data: function() {
    return {
      stages: stages,
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
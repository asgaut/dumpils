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
      </select>&nbsp;
      <label for="displayBW">Display bandwidth (Hz):</label>
      <input
        v-model.number="displayBW"
        type="number"
        id="displayBW"
        :min="500"
        :max="stages[stage].fs"
        :step="1000"
      />
    </div>
    <br />
    <SvgSpectrum
      :fs="stages[stage].fs"
      :url="`http://localhost:3344/spectrum?source=${source}&amp;stage=${stage}`"
      :displayBW="displayBW"
    />
  </div>
</template>

<script>
// @ is an alias to /src
import SvgSpectrum from "@/components/SvgSpectrum.vue";

const stages = {
  if: { name: "IF", fs: 1310720, defaultDisplayBW: 1e6 },
  lf: { name: "LF", fs: 81920, defaultDisplayBW: 2500 }
};

const spectrumGlobalState = {
  stages: stages,
  source: "loc",
  stage: "if",
  displayBW: Number(0)
};

export default {
  name: "Spectrum",
  data: function() {
    return spectrumGlobalState;
  },
  components: {
    SvgSpectrum
  },
  created: function() {
    if (this.displayBW === 0) {
      this.displayBW = this.stages["if"].defaultDisplayBW;
    }
  },
  watch: {
    stage: function(newStage) {
      if (this.displayBW > this.stages[newStage].fs) {
        this.displayBW = this.stages[newStage].defaultDisplayBW;
      }
    }
  }
};
</script>

<style>
</style>
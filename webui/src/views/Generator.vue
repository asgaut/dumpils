<template>
  <div class="root">
    <div class="controls">
      <h2>Main source</h2>
      <div>
        <div>
          <label for="ddmInput">DDM (%):</label>
          <input v-model.number="ddm" type="number" id="ddmInput" :min="-sdm" :max="sdm" step="0.1" />
          <input v-model="ddm" type="range" id="ddmRange" :min="-sdm" :max="sdm" step="0.1" />
        </div>
        <div>
          <label for="sdmInput">SDM (%):</label>
          <input v-model.number="sdm" type="number" id="sdmInput" :min="0" :max="200" step="1" />
          <input v-model="sdm" type="range" id="sdmRange" :min="0" :max="200" step="1" />
        </div>
        <div>
          <label for="rfInput">RF (% fullscale):</label>
          <input v-model.number="rf" type="number" id="rfInput" :min="0" :max="200" step="1" />
          <input v-model="rf" type="range" id="rfRange" :min="0" :max="200" step="1" />
        </div>
        <div>
          <label for="ifInput">Carrier offset (Hz):</label>
          <input
            v-model.number="carrierOffset"
            type="number"
            id="carrierOffset"
            :min="-50000"
            :max="50000"
            step="1000"
          />
        </div>
        <button v-on:click="ddm=0;sdm=40;rf=50;carrierOffset=0">Reset main source</button>
      </div>
      <div>
        <h2>Test source</h2>
        <div>
          <label for="ddmInput2">DDM (%):</label>
          <input
            v-model.number="ddm2"
            type="number"
            id="ddmInput2"
            :min="-sdm"
            :max="sdm"
            step="0.1"
          />
        </div>
        <div>
          <label for="sdmInput2">SDM (%):</label>
          <input v-model.number="sdm2" type="number" id="sdmInput2" :min="0" :max="200" step="1" />
        </div>
        <div>
          <label for="rfInput2">RF (% fullscale):</label>
          <input v-model.number="rf2" type="number" id="rfInput2" :min="0" :max="200" step="1" />
        </div>
        <div>
          <label for="carrierOffset2">Carrier offset (Hz):</label>
          <input
            v-model.number="carrierOffset2"
            type="number"
            id="carrierOffset2"
            :min="-50000"
            :max="50000"
            step="1000"
          />
        </div>
        <button v-on:click="ddm2=0;sdm2=40;rf2=0;carrierOffset2=0">Reset test source</button>
      </div>
      <br />
      <div class="error" v-if="overrange">Overrange ({{overrange}} samples)</div>
    </div>
  </div>
</template>

<script>
let fs = 1310720;
let sampleCount = Math.trunc(fs / 10);
let iqData = new Uint8ClampedArray(sampleCount * 2);

export default {
  name: "Generator",
  data: function() {
    return {
      ddm: 0,
      sdm: 40,
      rf: 50,
      carrierOffset: 0,
      ddm2: 0,
      sdm2: 40,
      rf2: 0,
      carrierOffset2: 0,
      overrange: 0
    };
  },
  created: function() {
    this.generate();
  },
  watch: {
    ddm: function() {
      this.generate();
    },
    sdm: function() {
      this.generate();
    },
    rf: function() {
      this.generate();
    },
    carrierOffset: function() {
      this.generate();
    },
    ddm2: function() {
      this.generate();
    },
    sdm2: function() {
      this.generate();
    },
    rf2: function() {
      this.generate();
    },
    carrierOffset2: function() {
      this.generate();
    }
  },
  methods: {
    generate: function() {
      let mod150 = (this.sdm / 2 + this.ddm / 2) / 100;
      let mod90 = (this.sdm / 2 - this.ddm / 2) / 100;
      let rf = this.rf / 100;
      let mod1502 = (this.sdm2 / 2 + this.ddm2 / 2) / 100;
      let mod902 = (this.sdm2 / 2 - this.ddm2 / 2) / 100;
      let rf2 = this.rf2 / 100;
      let nco_i = (freq, time) => Math.cos(2 * Math.PI * freq * time);
      let nco_q = (freq, time) => Math.sin(2 * Math.PI * freq * time);
      // prettier-ignore
      let baseband = time => rf * (1 + mod90*Math.sin(2*Math.PI*90*time) + mod150*Math.sin(2*Math.PI*150*time))
      // prettier-ignore
      let baseband2 = time => rf2 * (1 + mod902*Math.sin(2*Math.PI*90*time) + mod1502*Math.sin(2*Math.PI*150*time))
      performance.mark("markerNameA");
      iqData = new Uint8ClampedArray(sampleCount * 2);
      this.overrange = 0;
      for (let i = 0; i < sampleCount; i++) {
        let time = i / fs;
        // prettier-ignore
        let real = baseband(time) * nco_i(this.carrierOffset + Number(200e3), time);
        // prettier-ignore
        let imag = baseband(time) * nco_q(this.carrierOffset + Number(200e3), time);
        // prettier-ignore
        real += baseband2(time) * nco_i(this.carrierOffset2 + Number(200e3), time);
        // prettier-ignore
        imag += baseband2(time) * nco_q(this.carrierOffset2 + Number(200e3), time);
        iqData[2 * i] = 128 + 127 * real;
        iqData[2 * i + 1] = 128 + 127 * imag;
        if (real > 1) this.overrange++;
        else if (real < -1) this.overrange++;
        if (imag > 1) this.overrange++;
        else if (imag < -1) this.overrange++;
      }
      performance.mark("markerNameB");
      performance.measure("measure a to b", "markerNameA", "markerNameB");
      let duration = performance.getEntriesByType("measure")[0].duration; // about 50 ms on my machine
      performance.clearMarks();
      performance.clearMeasures();
      console.log("generated", iqData.length, "samples in", duration, "ms");
      /*var file = new Blob([iqData], {
        type: "application/octet-binary"
      });
      var a = document.createElement("a"),
        url = URL.createObjectURL(file);
      a.href = url;
      a.download = "iqdata.iq";
      document.body.appendChild(a);
      a.click();*/
      this.upload();
    },
    upload: function() {
      fetch("http://localhost:3344/samples?source=loc", {
        method: "PUT", // *GET, POST, PUT, DELETE, etc.
        mode: "cors", // no-cors, *cors, same-origin
        cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
        credentials: "same-origin", // include, *same-origin, omit
        headers: {
          "Content-Type": "application/octet-binary"
        },
        redirect: "follow", // manual, *follow, error
        referrerPolicy: "no-referrer", // no-referrer, *client
        body: iqData // body data type must match "Content-Type" header
      })
        .then(response => {
          console.log("New sim data set, response: ", response.status);
        })
        .catch(e => {
          console.log(`error setting sim data ${e}`);
        });
    }
  }
};
</script>

<style scoped>
.error {
  color: red;
  font-weight: bold;
}
</style>
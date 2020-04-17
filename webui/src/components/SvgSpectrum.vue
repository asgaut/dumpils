<template>
  <div ref="spectrumdiv">
    <svg
      viewBox="-1000 0 2000 1000"
      xmlns="http://www.w3.org/2000/svg"
      xmlns:svg="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id="linear" x1="0" y1="0" x2="0" y2="1000" gradientUnits="userSpaceOnUse">
          <stop offset="0%" stop-color="#c00" />
          <stop offset="100%" stop-color="#007" />
        </linearGradient>
      </defs>
      <g>
        <title>Spectrum</title>
        <path id="frame" fill="#222" d="M-1000 -1000 v2000 h2000 v-2000 " />
        <g v-for="(db, index) in power" v-bind:key="index">
          <rect
            :x="2000*(index<power.length/2?index:(index-power.length))/power.length"
            :y="1000-(db-dbMin)/(dbMax-dbMin)*1000"
            :width="2000/power.length"
            :height="(db-dbMin)/(dbMax-dbMin)*1000"
            fill="url(#linear)"
          >
            <title>{{db}}</title>
          </rect>
        </g>
      </g>
    </svg>
  </div>
</template>

<script>
export default {
  name: "SvgSpectrum",
  components: {},
  data: function() {
    return {
      dbMin: Number(-100),
      dbMax: 10, // max is 3.02 dBFS
      power: [0, 0]
    };
  },
  props: {
    url: {
      type: String,
      default: ""
    }
  },
  mounted() {
    this.updateData();
  },
  beforeDestroy() {
    clearInterval(this.timerData);
  },
  methods: {
    updateData: function() {
      let a = this.url;
      //this.displaySampleData = false;
      // Stop timer when doing network IO
      clearInterval(this.timerData);
      if (this.url.length == 0) {
        console.error("no url set");
        return;
      }
      fetch(this.url, {
        method: "GET", // *GET, POST, PUT, DELETE, etc.
        mode: "cors", // no-cors, *cors, same-origin
        cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
        credentials: "same-origin", // include, *same-origin, omit
        headers: {
          "Content-Type": "application/json"
        }
      })
        .then(response => response.json())
        .then(json => {
          let minBinWidth = 3;
          let downsample = 1;
          let clientWidth = 500;
          if (this.$refs.spectrumdiv) {
            clientWidth = this.$refs.spectrumdiv.clientWidth;
          }
          if (json.length > clientWidth / minBinWidth) {
            downsample = Math.trunc((json.length / clientWidth) * minBinWidth);
          }
          let reduced = [];
          let count = 0;
          let max = -Infinity;
          // Normalizes data in dB to be in range dbMin to dbMax
          json.forEach(item => {
            count++;
            let v = 20 * Math.log10(item);
            if (v < this.dBMin) v = this.dbMin;
            else if (v > this.dbMax) v = this.dbMax;
            else if (isNaN(v)) v = this.dbMin;
            if (v > max) max = v;
            if (count === downsample) {
              reduced.push(max);
              max = this.dbMin;
              count = 0;
            }
          });
          this.power = reduced;
          // Restart timer since fetch was successful
          this.timerData = setInterval(this.updateData, 500);
        })
        .catch(function(e) {
          console.error(`error fetching data from ${a}: ${e}`);
        });
    }
  }
};
</script>

<style>
</style>
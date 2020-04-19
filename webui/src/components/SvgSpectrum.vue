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
            <title>{{binLabel(db, index)}}</title>
          </rect>
        </g>
      </g>
    </svg>
    Span: {{fs}} Hz
    RBW: {{rbw}} Hz
    Sweep time: {{sweeptime}} s
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
      downsample: 0,
      power: [0, 0],
      specFull: [],
      specZoom: [],
      rbw: 0,
      sweeptime: 0
    };
  },
  props: {
    url: {
      type: String,
      default: ""
    },
    fs: Number(0),
    displayBW: Number(0)
  },
  mounted() {
    this.updateData();
  },
  beforeDestroy() {
    clearInterval(this.timerData);
  },
  watch: {
    displayBW: function() {
      this.redraw();
    },
    url: function() {
      this.updateData();
    },
    fs: function() {
      this.redraw();
    }
  },
  computed: {},
  methods: {
    binLabel: function(db, index) {
      if (index > 0) {
        index =
          index > this.power.length / 2 ? index - this.power.length : index;
        return `${index * this.downsample * this.rbw} Hz / ${db.toFixed(1)} dB`;
      }
      return `DC: ${db.toFixed(1)} dB`;
    },
    redraw: function() {
      this.sweeptime = this.specFull.length / this.fs;
      this.rbw = 1 / this.sweeptime;
      let showBins = Math.trunc(this.displayBW / this.rbw / 2);
      // keep 'showBins*2' samples of the lowest frequencies
      this.specZoom = [
        ...this.specFull.slice(0, showBins),
        ...this.specFull.slice(-showBins)
      ];
      let minBinWidth = 3;
      let clientWidth = 500;
      if (this.$refs.spectrumdiv) {
        clientWidth = this.$refs.spectrumdiv.clientWidth;
      }
      this.downsample = 1;
      if (this.specZoom.length > clientWidth / minBinWidth) {
        this.downsample = Math.trunc(
          (this.specZoom.length / clientWidth) * minBinWidth
        );
      }
      let reduced = [];
      let count = 0;
      let max = -Infinity;
      // Normalizes data in dB to be in range dbMin to dbMax
      this.specZoom.forEach(item => {
        count++;
        let v = 20 * Math.log10(item);
        if (v < this.dBMin) v = this.dbMin;
        else if (v > this.dbMax) v = this.dbMax;
        else if (isNaN(v)) v = this.dbMin;
        if (v > max) max = v;
        if (count === this.downsample) {
          reduced.push(max);
          max = this.dbMin;
          count = 0;
        }
      });
      this.power = reduced;
    },
    updateData: function() {
      let a = this.url;
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
          this.specFull = json;
          this.redraw();
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
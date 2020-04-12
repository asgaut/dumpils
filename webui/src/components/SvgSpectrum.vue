<template>
  <div ref="spectrumdiv">
    <svg viewBox="-1000 0 2000 1000"
      xmlns="http://www.w3.org/2000/svg"
      xmlns:svg="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient id="linear" x1="0" y1="0" x2="0" y2="800" gradientUnits="userSpaceOnUse">
          <stop offset="0%"   stop-color="#c00"/>
          <stop offset="100%" stop-color="#007"/>
        </linearGradient>
      </defs>
      <g>
        <title>Spectrum</title>
        <path id="frame" fill="#222" d="M-1000 -1000 v2000 h2000 v-2000 " />
        <g v-for="(f, index) in power" v-bind:key="index">
          <rect :x="index*binWidth-1000" :y="Math.abs(-(f+10)/totalmax)*1000" :width="binWidth" :height="Math.abs(1+(f+10)/totalmax)*1000" fill="url(#linear)">
            <title>{{f}}</title>
          </rect>
        </g>
      </g>
    </svg>
  </div>
</template>

<script>
export default {
  name: "SvgSpectrum",
  components: {
  },
  data: function() {
    return {
      totalmax: 100,
      power: [0, 0]
    };
  },
  props: {
    url: {
      type: String,
      default: ""
    },
  },
  computed: {
    binWidth: function() {
      return 2000/this.power.length;
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
        method: 'GET', // *GET, POST, PUT, DELETE, etc.
        mode: 'cors', // no-cors, *cors, same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
          'Content-Type': 'application/json',
        }}).then(response => response.json())
        .then(json => {
          let minBinWidth = 3;
          let downsample = 1;
          let clientWidth = 500;
          //let downsample = Math.trunc(json.length / this.$refs.spectrum2.clientWidth * minBinWidth);
          if (this.$refs.spectrumdiv) {
            clientWidth = this.$refs.spectrumdiv.clientWidth
          }
          if (json.length > clientWidth/minBinWidth) {
            downsample = Math.trunc(json.length / clientWidth * minBinWidth);
          }
          let reduced = [];
          let max = -90; // dBFS
          let count = 0;
          // Normalizes data in dB to be in range -90 to 10
          json.forEach(item => {
            count++;
            let v = 20*Math.log10(item)
            if (v>10) v = 10;
            if (v>max) max = v;
            if (count === downsample) {
              reduced.push(max);
              //if (max > totalmax) totalmax = max;
              max = -90; // dBFS
              count = 0;
            }
          });
          //this.totalmax = totalmax;
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
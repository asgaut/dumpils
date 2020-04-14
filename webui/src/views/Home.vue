<template>
  <div>
    <div class="main">
      <div class="cdi" @click="cdiClick">
        <CDI
          :nav-current="navCurrent"
          :nav-flag="navFlag"
          :gs-current="gsCurrent"
          :gs-flag="gsFlag"
        />
        <div
          style="text-align: center; font-size: small; cursor: pointer;"
        >{{`Click here to ${showControls?'hide':'show'} the control panel.`}}</div>
      </div>
      <div class="controls" v-show="showControls">
        <label for="channel">Channel:</label>
        <select id="channel" v-model="selectedChannel">
          <option disabled value>Please select one</option>
          <option
            v-for="f in channels"
            v-bind:key="f.name"
            v-bind:value="f.name"
          >{{`${f.name} (${f.loc} / ${f.gp})`}}</option>
        </select>
        <br />
        <input type="checkbox" id="checkbox" v-model="showYChannels" />
        <label for="checkbox">Show Y channels</label>
        <div class="meas">
          <div v-if="measurements['loc']" :set="m = measurements['loc']">
            LOC:
            <br />
            <table>
              <tr>
                <td>DDM:</td>
                <td style="text-align:right;">{{(navCurrent).toFixed(1)}}&nbsp;</td>
                <td>µA</td>
              </tr>
              <tr>
                <td>SDM:</td>
                <td style="text-align:right;">{{m.sdm.toFixed(1)}}</td>
                <td>%</td>
              </tr>
              <tr>
                <td>RF:</td>
                <td style="text-align:right;">{{m.rf.toFixed(1)}}</td>
                <td>dBFS</td>
              </tr>
            </table>
          </div>
          <div v-else>LOC: No data</div>
          <div v-if="measurements['gp']" :set="m = measurements['gp']">
            GP:
            <br />
            <table>
              <tr>
                <td>DDM:</td>
                <td style="text-align:right;">{{(gsCurrent).toFixed(1)}}&nbsp;</td>
                <td>µA</td>
              </tr>
              <tr>
                <td>SDM:</td>
                <td style="text-align:right;">{{m.sdm.toFixed(1)}}</td>
                <td>%</td>
              </tr>
              <tr>
                <td>RF:</td>
                <td style="text-align:right;">{{m.rf.toFixed(1)}}</td>
                <td>dBFS</td>
              </tr>
            </table>
          </div>
          <div v-else>GP: No data</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
// @ is an alias to /src
import CDI from "@/components/CDI.vue";
import allChannels from "@/channels.js";

export default {
  name: "Home",
  components: {
    CDI
  },
  data: function() {
    return {
      selectedChannel: false,
      showYChannels: false,
      showControls: true,
      measurements: {}
    };
  },
  watch: {
    selectedChannel: function(val) {
      let channel = allChannels.filter(i => i.name === val);
      if (channel.length !== 1) console.error(val, "not found");
      console.log("Sending new channel: ", channel[0]);
      // Default options are marked with *
      fetch("http://localhost:3344/channel", {
        method: "PUT", // *GET, POST, PUT, DELETE, etc.
        mode: "cors", // no-cors, *cors, same-origin
        cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
        credentials: "same-origin", // include, *same-origin, omit
        headers: {
          "Content-Type": "application/json"
        },
        redirect: "follow", // manual, *follow, error
        referrerPolicy: "no-referrer", // no-referrer, *client
        body: JSON.stringify(channel[0]) // body data type must match "Content-Type" header
      })
        .then(response => response.json())
        .then(json => {
          console.log("New channel set, response: ", json);
        })
        .catch(e => {
          console.log(`error setting channel ${e}`);
        });
    }
  },
  computed: {
    channels: function() {
      // Return items to show in the channel selection dropdown list
      return allChannels.filter(
        i => (this.showYChannels || i.name.endsWith("X")) && i.loc != undefined
      );
    },
    navCurrent: function() {
      return this.measurements["loc"]
        ? (this.measurements["loc"].ddm * 150) / 15.5
        : 0;
    },
    navFlag: function() {
      let sdm_alarm =
        this.measurements["loc"]?.sdm < 30 ||
        this.measurements["loc"]?.sdm > 50;
      return this.measurements["loc"] == undefined || sdm_alarm;
    },
    gsCurrent: function() {
      return this.measurements["gp"]
        ? (this.measurements["gp"].ddm * 150) / 17.5
        : 0;
    },
    gsFlag: function() {
      let sdm_alarm =
        this.measurements["gp"]?.sdm < 70 || this.measurements["gp"]?.sdm > 90;
      return this.measurements["gp"] == undefined || sdm_alarm;
    }
  },
  mounted() {
    console.log("starting measurement update");
    this.updateData();
  },
  beforeDestroy() {
    clearInterval(this.timerData);
  },
  methods: {
    cdiClick: function() {
      this.showControls = !this.showControls;
    },
    updateData: function() {
      let url = "http://localhost:3344/measurements";
      clearInterval(this.timerData);
      fetch(url)
        .then(response => response.json())
        .then(json => {
          this.measurements = json;
          // Restart timer since fetch was successful
          this.timerData = setInterval(this.updateData, 500);
        })
        .catch(e => {
          this.measurements = {};
          console.error(`error fetching data from ${url}: ${e}`);
        });
    }
  }
};
</script>

<style scoped>
/* https://www.youtube.com/watch?v=JJSoEo8JSnc */
.top {
  border: 2px #666 solid;
  padding: 7px;
}

.main div {
  padding: 10px;
}
.cdi {
  flex: 2;
}
.controls {
  flex: 1;
}
.meas {
  font-size: large;
  text-align: left;
}
@media (min-width: 600px) {
  .main {
    display: flex;
    border: 1px #ccc solid;
    justify-content: center;
  }
}
</style>

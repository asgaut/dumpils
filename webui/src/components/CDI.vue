<template>
  <svg
    :width="width"
    :height="height"
    style="max-height:500px"
    viewBox="-1000 -1000 2000 2000"
    xmlns="http://www.w3.org/2000/svg"
    xmlns:svg="http://www.w3.org/2000/svg"
  >
    <defs>
      <symbol id="gs_tick">
        <g transform="translate(100 100)">
          <circle fill="white" r="30" />
          <rect fill="white" x="-50" y="-7" height="14" width="100" />
        </g>
      </symbol>
      <symbol id="nav_tick">
        <g transform="translate(100 100)">
          <circle fill="white" r="30" />
          <rect fill="white" x="-7" y="-50" height="100" width="14" />
        </g>
      </symbol>
      <mask id="dialmask">
        <circle cx="0" cy="0" r="935" style="fill: #ffffff" />
      </mask>
    </defs>
    <g>
      <title>Course Deviation Indicator</title>
      <path
        id="frame"
        fill="#121212"
        d="M 0 -1000
          h 800 l 200 200 v 800 v 800 l -200 200 h -800
          h -800 l -200 -200 v -800 -800 l 200 -200"
      />
      <circle id="bezel-inner" cx="0" cy="0" r="970" stroke="#888" fill="black" stroke-width="50" />
      <circle id="bezel-outer" cx="0" cy="0" r="980" stroke="#eee" fill="none" stroke-width="20" />
      <g id="dots" transform="translate(-100 -100)">
        <use href="#gs_tick" x="0" y="-800" />
        <use href="#gs_tick" x="0" y="-640" />
        <use href="#gs_tick" x="0" y="-480" />
        <use href="#gs_tick" x="0" y="-320" />
        <use href="#gs_tick" x="0" y="-160" />
        <use href="#gs_tick" x="0" y="160" />
        <use href="#gs_tick" x="0" y="320" />
        <use href="#gs_tick" x="0" y="480" />
        <use href="#gs_tick" x="0" y="640" />
        <use href="#gs_tick" x="0" y="800" />
        <use href="#nav_tick" x="-800" y="0" />
        <use href="#nav_tick" x="-640" y="0" />
        <use href="#nav_tick" x="-480" y="0" />
        <use href="#nav_tick" x="-320" y="0" />
        <use href="#nav_tick" x="-160" y="0" />
        <use href="#nav_tick" x="160" y="0" />
        <use href="#nav_tick" x="320" y="0" />
        <use href="#nav_tick" x="480" y="0" />
        <use href="#nav_tick" x="640" y="0" />
        <use href="#nav_tick" x="800" y="0" />
        <circle cx="100" cy="100" r="80" stroke="white" stroke-width="20" />
      </g>
      <g id="nav-flag" v-show="navFlag">
        <rect x="200" y="-600" fill="red" width="170" height="360" />
        <text
          x="235"
          y="-480"
          style="font-size:130px;fill:black;font-weight:bold;font-family:sans-serif;"
        >N</text>
        <text
          x="235"
          y="-370"
          style="font-size:130px;fill:black;font-weight:bold;font-family:sans-serif;"
        >A</text>
        <text
          x="237"
          y="-260"
          style="font-size:130px;fill:black;font-weight:bold;font-family:sans-serif;"
        >V</text>
      </g>
      <g id="gs-flag" v-show="gsFlag">
        <rect x="-800" y="-300" fill="red" width="250" height="170" />
        <text
          x="-770"
          y="-175"
          style="font-size:130px;fill:black;font-weight:bold;font-family:sans-serif;"
        >GS</text>
      </g>
      <g mask="url(#dialmask)">
        <rect
          id="nav-needle"
          :transform="`translate(${-navDeflection/150*800},0)`"
          fill="white"
          x="-15"
          y="-800"
          width="30"
          height="1600"
        />
        <rect
          id="gs-needle"
          :transform="`translate(0,${-gsDeflection/150*800})`"
          fill="white"
          x="-800"
          y="-15"
          width="1600"
          height="30"
        />
      </g>
    </g>
  </svg>
</template>

<script>
export default {
  name: "CDI",
  props: {
    width: {
      type: String
    },
    height: {
      type: String
    },
    navCurrent: Number,
    gsCurrent: Number,
    navFlag: {
      type: Boolean,
      default: true
    },
    gsFlag: {
      type: Boolean,
      default: true
    }
  },
  computed: {
    navDeflection: function() {
      return this.navCurrent < -150
        ? -150
        : this.navCurrent > 150
        ? 150
        : this.navCurrent;
    },
    gsDeflection: function() {
      return this.gsCurrent < -150
        ? -150
        : this.gsCurrent > 150
        ? 150
        : this.gsCurrent;
    }
  }
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>

import Vue from "vue";
import VueRouter from "vue-router";
import Home from "../views/Home.vue";
import Spectrum from "../views/Spectrum.vue";
import Generator from "../views/Generator.vue";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "Home",
    component: Home
  },
  {
    path: "/spectrum",
    name: "Spectrum",
    component: Spectrum
  },
  {
    path: "/generator",
    name: "Generator",
    component: Generator
  }, {
    path: "/about",
    name: "About",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/About.vue")
  }
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes
});

export default router;

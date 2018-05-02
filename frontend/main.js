import World from './World';
import Gear from './actors/Gear';
import Index from './layers/Index';
import Page from './layers/Page';

/** @type {World} */
let world = null;

function main() {
  /** @type {HTMLCanvasElement} */
  const canvas = document.getElementById('background');
  if(!canvas) {
    document.body.innerHTML='<h1>No canvas</h1>';
    return;
  }
  world = World.fromCanvas(canvas);
  if(!world) {
    document.body.innerHTML='<h1>WebGL not supported</h1>';
    return;
  }
  world.start();

  open(location.pathname);
}

/**
 * @param {string} pathName 
 */
function open(pathName) {
  if(pathName == '/') {
    const index = new Index(world);
    world.pushLayer(index);
  }else if(pathName.startsWith('/about-us/')){

  }else{
    const url = `/moment${pathName}`;
    const content = fetch(url).then(resp => resp.text());
    world.pushLayer(new Index(world));
    world.pushLayer(new Page(world, content));
  }
}

document.addEventListener('DOMContentLoaded', function() {
  main();
}, false);
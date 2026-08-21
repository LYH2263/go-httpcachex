async function load(){
  document.getElementById('e').textContent=JSON.stringify(await (await fetch('/api/entries')).json(),null,2);
  document.getElementById('s').textContent=JSON.stringify(await (await fetch('/api/stats')).json(),null,2);
}
document.getElementById('r').onclick=load; load(); setInterval(load,3000);

import http from 'k6/http';
import { sleep } from 'k6';
import { Trend } from 'k6/metrics';

/* ─── нагрузочный профиль ─── */
export const options = {
  stages: [
    { duration: '30s', target: 500 },    // Быстрый старт с 500 VU
    { duration: '1m',  target: 1500 },   // Увеличение до 1500 VU за 1 минуту
    { duration: '2m',  target: 3000 },   // Увеличение до 3000 VU за 2 минуты
    { duration: '1m',  target: 4000 },   // Увеличение до пика 4000 VU
    { duration: '30s', target: 0 },      // Плавный спад
  ],
  thresholds: {
    http_req_failed: ['rate<0.1'],
    e2e_time:        ['p(95)<60000'],
    instant_rps:     ['p(99)>0'],      // просто чтобы метрика попала в summary
  },
  summaryTrendStats: ['avg','min','med','max','p(90)','p(95)','p(99)'],
};

/* ─── кастомные метрики ─── */
const e2eTime    = new Trend('e2e_time');
const createTime = new Trend('create_time');
const pollTime   = new Trend('poll_time');
const instantRps = new Trend('instant_rps', /*saveSamples=*/true);

/* ─── глобалки ─── */
const API = __ENV.API_URL || 'http://go_init_manager:60013/graphql';
let   tickStart = Date.now();
let   reqCnt    = 0;

/* ─── утилиты ─── */
function gql(q, v) {
  return http.post(
    API,
    JSON.stringify({ query: q, variables: v }),
    { headers: { 'Content-Type': 'application/json' },
      timeout: '10s',
    }
  );
}
function randSvc() { return 'svc-' + Math.random().toString(36).slice(2,10); }

/* ─── основной сценарий ─── */
export default function () {
  /* секундные окна RPS */
  const now = Date.now();
  if (now - tickStart >= 1000) {
    instantRps.add(reqCnt);          // 1 точка = req-per-sec
    reqCnt   = 0;
    tickStart = now;
  }

  const t0 = Date.now();

  /* 1. createTemplate */
  const c0 = Date.now();
  const cRes = gql(`
    mutation($in:CreateTemplateInput!){
      createTemplate(input:$in){ template{ id status zipUrl } }
    }`, { in: {
      name: randSvc(),
      endpoints:[{protocol:'GRPC',role:'SERVER'}],
      database:{type:'POSTGRESQL',ddl:'CREATE TABLE t(id int);'},
      docker:{registry:'docker.io',imageName:'demo'}
    }});
  reqCnt++;
  createTime.add(Date.now()-c0);
  if (cRes.status!==200) return;
  const tpl=cRes.json('data.createTemplate.template');
  if(!tpl?.id) return;

  /* 2. poll ≤30 с */
  const deadline = Date.now()+30000;
  let status=tpl.status;
  const p0 = Date.now();
  while(Date.now()<deadline && !['COMPLETED','FAILED'].includes(status)){
    sleep(0.05);  // Уменьшено время задержки с 0.1 до 0.05 для увеличения частоты опроса
    const pRes=gql(`query($id:ID!){getTemplate(id:$id){template{status zipUrl}}}`,{id:tpl.id});
    reqCnt++;
    if(pRes.status===200){
      status=pRes.json('data.getTemplate.template.status')||status;
    }
  }
  pollTime.add(Date.now()-p0);

  /* 3. e2e */
  e2eTime.add(Date.now()-t0,{final:status});
}

/* ─── сохранение итогов ─── */
export function handleSummary(data){
  return {
    'k6_summary.json': JSON.stringify(data,null,2), // агрегаты
    'k6_metrics.json': JSON.stringify({
      points: data.metrics.instant_rps.values   // properly formatted JSON object
    })
  };
}

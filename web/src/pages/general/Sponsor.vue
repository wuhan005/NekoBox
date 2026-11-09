<template>
  <div class="uk-card uk-card-default uk-card-body">
    <div class="uk-child-width-expand@s" uk-grid>
      <div>
        <h2>给 NekoBox 打钱</h2>
        <p>NekoBox 创立于 2020 年疫情初期，一开始是我疫情在家花了三天时间随手写得一个匿名留言箱。</p>
        <p>目前 NekoBox 的服务器开销还是我一个人承担，如果你也想为我分担一些，欢迎给 NekoBox 打钱！👏🏻</p>
      </div>
      <div>
        <div class="sponsor">
          <div class="alipay popup" style="">
            <div class="head">谢谢老板 Thanks♪(･ω･)ﾉ</div>
            <div class="qrcode">
              <div class="qrcode-li alipay"><img
                  src="https://img13.360buyimg.com/ceco/jfs/t1/2848/30/22668/22212/66b6ee87F15f75426/9f8969181dd1f244.png"
                  referrerpolicy="no-referrer">
                <noscript><img
                    src="https://img13.360buyimg.com/ceco/jfs/t1/2848/30/22668/22212/66b6ee87F15f75426/9f8969181dd1f244.png"
                    referrerpolicy="no-referrer"/></noscript>
              </div>
            </div>
            <ul class="platform">
              <li class="icon-alipay active" data-bg-color="#00a2ea" data-thanks="谢谢老板 Thanks♪(･ω･)ﾉ"></li>
            </ul>
          </div>
        </div>
      </div>
    </div>
    <div uk-alert class="uk-text-small">感谢以下小伙伴对 NekoBox 的支持！</div>
    <table class="uk-table uk-table-small uk-table-divider uk-text-small">
      <tbody v-html="sponsorHtml"></tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted} from "vue";
import {getSponsors} from "@/utils/feeds.ts";

const sponsorHtml = ref<string>('')

onMounted(() => {
  getSponsors().then(res => {
    sponsorHtml.value = res.data[0].content
  })
})
</script>

<style scoped>
.sponsor {
  position: relative;
}

.sponsor > .popup {
  width: 280px;
  background: #eee;
  color: #fff;
  box-shadow: 0 0 5px 0 #ccc;
  border-radius: 5px;
  line-height: 0;
}

.sponsor > .alipay {
  background: #00a2ea;
}

.sponsor > .popup > .head {
  text-align: center;
  padding-top: 30px;
}

.sponsor .qrcode {
  text-align: center;
}

.sponsor .qrcode img {
  margin: 30px auto;
  width: 80%;
}

.sponsor ul.platform {
  padding: 0 !important;
  margin: 0 !important;
  text-align: center;
  background: #FFF;
  border-bottom-left-radius: 5px;
  border-bottom-right-radius: 5px;
}

.sponsor ul.platform li {
  display: inline-block;
  cursor: pointer;
  margin: 5px;
  filter: grayscale(100%);
}

.sponsor ul.platform li.active {
  font-weight: 600;
  filter: grayscale(0%);
}

.sponsor ul.platform li.icon-alipay {
  background: url(https://img10.360buyimg.com/ceco/jfs/t1/188386/33/41420/3658/66b6ecf9F03bfd7eb/781b0e5e0a6db211.png) no-repeat center center;
  width: 24px;
  height: 24px;
  background-size: cover;
}
</style>
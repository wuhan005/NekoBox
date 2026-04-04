<template>
  <div id="qrcode-modal" class="uk-flex-top" uk-modal>
    <div class="uk-modal-dialog uk-modal-body uk-margin-auto-vertical" style="max-width: 440px !important">
      <button class="uk-modal-close-default" type="button" uk-close></button>
      <div class="qrcode">
        <div class="background">
          <div class="container">
            <img alt="QR Code" class="code" :src="qrcodeVal"/>
            <img alt="avatar" class="avatar" :src="props.avatar"/>
          </div>
          <div class="desc">
            {{ props.name }} 的提问箱
            <p>扫一扫二维码 向我提问</p>
          </div>
        </div>
      </div>
      <div class="uk-text-center uk-text-small uk-text-muted uk-margin-top">截图保存二维码</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {defineProps} from 'vue';
import {computed} from 'vue';
import {useQRCode} from '@vueuse/integrations/useQRCode'

const props = defineProps({
  name: String,
  avatar: String,
  qrcode: String
})

const qrcodeVal = useQRCode(computed(() => props.qrcode || ''), {
  width: 400,
  height: 400,
  colorDark: "#000000",
  colorLight: "#ffffff",
  correctLevel: 'H'
})
</script>

<style scoped lang="less">
.qrcode {
  width: 100%;
  padding-top: 100%;
  position: relative;

  .background {
    position: absolute;
    max-width: 400px;
    max-height: 400px;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-image: url('https://box-user-assets.n3ko.cc/public/qrcode_bg');
    background-size: 100% 100%;
    background-repeat: no-repeat;
    user-select: none;
    -webkit-user-select: none;

    .container {
      position: absolute;
      top: 16.5%;
      left: 26.5%;
      width: 47%;
      height: 47%;

      .code {
        width: 100%;
        height: 100%;
        border: 8% solid #ffffff;
        box-sizing: border-box;
      }

      .avatar {
        position: absolute;
        width: 25%;
        height: 25%;
        top: 38%;
        left: 38%;
        background: #fff;
        border: 2px solid #ffffff;
        border-radius: 5%;
      }
    }

    .desc {
      font-family: lucida grande, helvetica neue, Helvetica, Arial, Verdana, pingfang sc, STHeiti, microsoft yahei, SimSun, sans-serif;
      position: absolute;
      top: 73%;
      width: 100%;
      font-size: 140%;
      text-align: center;
      color: #ffffff;

      p {
        font-size: 60%;
        margin-top: 3%;
      }
    }
  }
}
</style>
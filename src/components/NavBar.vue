<template>
  <aside
    ref="navRoot"
    :class="['menu', 'main-nav', 'has-background-primary', 'soft-if-dark', { resizing: isResizing }]"
    :style="sidebarStyle"
  >
    <div class="top-sheet">
      <router-link to="/" v-slot="{ navigate }">
        <img alt="logo" class="logo" src="@/assets/logo-white.svg" @click="navigate" />
      </router-link>

      <!-- non-workspace general routes -->
      <template v-if="routeIsDashboard">
        <p class="menu-label">General</p>
        <ul class="menu-list">
          <li>
            <router-link to="/">
              <FontAwesomeIcon class="mr-1" :icon="faTableCells" size="sm" />
              Dashboard</router-link
            >
          </li>
        </ul>

        <p class="menu-label">About</p>
        <ul class="menu-list">
          <li>
            <router-link to="/about">
              <FontAwesomeIcon class="mr-1" :icon="faCircleInfo" size="sm" />
              About</router-link
            >
          </li>
          <li>
            <a href="https://github.com/pulsejet/ownly" target="_blank">
              <FontAwesomeIcon class="mr-1" :icon="faGithub" size="sm" />
              GitHub
            </a>
          </li>
        </ul>
      </template>

      <template v-if="routeIsWorkspace">
        <p class="menu-label">Projects</p>
        <ul class="menu-list">
          <li v-for="proj in projects" :key="proj.uuid" class="project-item">
            <router-link :to="linkProject(proj)" class="project-link">
              <div class="link-inner project-link-inner">
                <span class="project-icon-shell">
                  <FontAwesomeIcon :icon="faLayerGroup" size="sm" />
                </span>
                <span class="project-name">{{ proj.name }}</span>
              </div>

              <ProjectTreeMenu
                v-if="activeProjectName === proj.uuid"
                class="link-button"
                :allow-new="true"
                :allow-delete="false"
                @new-file="projectTree?.[0]?.newHere('file', $event)"
                @new-folder="projectTree?.[0]?.newHere('folder')"
                @import="projectTree?.[0]?.importHere()"
                @import-zip="projectTree?.[0]?.importZipHere()"
                @export="projectTree?.[0]?.executeExport(null)"
              />
            </router-link>

            <Transition name="project-tree-collapse">
              <ProjectTree
                v-if="activeProjectName == proj.uuid"
                class="outermost"
                ref="projectTree"
                :project="proj"
                :files="projectFiles"
              />
            </Transition>
          </li>

          <li>
            <a @click="showProjectModal = true">
              <FontAwesomeIcon class="mr-1" :icon="faPlus" size="sm" />
              Add project
            </a>
          </li>
        </ul>

        <p class="menu-label">Discussion</p>
        <ul class="menu-list">
          <li v-for="chan in chatChannels" :key="chan.uuid">
            <router-link :to="linkDiscuss(chan)">
              <FontAwesomeIcon class="mr-1" :icon="faHashtag" size="sm" />
              {{ chan.name }}
            </router-link>
          </li>
          <li>
            <a @click="showChannelModal = true">
              <FontAwesomeIcon class="mr-1" :icon="faPlus" size="sm" />
              Add channel
            </a>
          </li>
        </ul>

        <p class="menu-label">AI Agents</p>
        <ul class="menu-list">
          <li>
            <a @click="showAgentModal = true">
              <FontAwesomeIcon class="mr-1" :icon="faRobot" size="sm" />
              Manage agents
            </a>
          </li>
        </ul>

        <p class="menu-label">Workspace</p>
        <ul class="menu-list">
          <li>
            <a @click="showInviteModal = true">
              <FontAwesomeIcon class="mr-1" :icon="faPlus" size="sm" />
              Invite people

              <FontAwesomeIcon v-show="showNotifBubble" class="mr-1" :icon="faCircleExclamation" size="sm"></FontAwesomeIcon>
            </a>
          </li>
        </ul>
      </template>

      <p class="menu-label">v{{ buildVersion() }}</p>
    </div>

    <div class="bottom-sheet">
      <div class="id-share">
        <a @click="showIdentity = true">
          <FontAwesomeIcon class="mr-1" :icon="faQrcode" size="sm" />
          Share your Identity
        </a>
      </div>

      <div class="connection">
        <template v-if="connState.connected">
          <FontAwesomeIcon class="mr-1" :icon="faWifi" size="sm" />
          {{ connState.router }}
        </template>
        <template v-else>
          <FontAwesomeIcon class="mr-1" :icon="faGhost" size="sm" />
          Offline
        </template>
      </div>
    </div>

    <div class="sidebar-resizer" @pointerdown.prevent="startSidebarResize"></div>

    <AddChannelModal :show="showChannelModal" @close="showChannelModal = false" />
    <AddProjectModal :show="showProjectModal" @close="showProjectModal = false" />
    <InvitePeopleModal :show="showInviteModal" @close="showInviteModal = false" />
    <QRIdentityModal :show="showIdentity" @close="showIdentity = false" />

    <ModalComponent :show="showAgentModal" @close="showAgentModal = false">
      <AgentBrowser />
    </ModalComponent>
  </aside>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, useTemplateRef } from 'vue';
import { useRoute } from 'vue-router';
//import { useRouter } from 'vue-router';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import {
  faLayerGroup,
  faPlus,
  faHashtag,
  faWifi,
  faGhost,
  faTableCells,
  faQrcode,
  faCircleInfo,
  faRobot,
  faCircleExclamation,
} from '@fortawesome/free-solid-svg-icons';
import { faGithub } from '@fortawesome/free-brands-svg-icons';

import ProjectTree from './ProjectTree.vue';
import ProjectTreeMenu from './ProjectTreeMenu.vue';
import AddChannelModal from './AddChannelModal.vue';
import AddProjectModal from './AddProjectModal.vue';
import AgentBrowser from './AgentBrowser.vue';
import ModalComponent from './ModalComponent.vue';

import { GlobalBus } from '@/services/event-bus';
import { Toast } from '@/utils/toast';
//import { Workspace } from '@/services/workspace';

import type { IChatChannel, IProject, IProjectFile } from '@/services/types';
import InvitePeopleModal from './InvitePeopleModal.vue';
import QRIdentityModal from './QRIdentityModal.vue';

const route = useRoute();
//const router = useRouter();
const routeIsDashboard = computed(() =>
  ['dashboard', 'join', 'about'].includes(String(route.name)),
);
const routeIsWorkspace = computed(() =>
  ['space-home', 'project', 'discuss', 'project-file'].includes(String(route.name)),
);

const showChannelModal = ref(false);
const showProjectModal = ref(false);
const showInviteModal = ref(false);
const showIdentity = ref(false);
const showAgentModal = ref(false);

// vue-tsc chokes on this type inference
const projectTree = useTemplateRef<Array<InstanceType<typeof ProjectTree>>>('projectTree');

const channels = ref([] as IChatChannel[]);

// Use channels directly as chatChannels since they're now separate
const chatChannels = computed(() => channels.value);

const projects = ref([] as IProject[]);
const activeProjectName = ref(null as string | null);
const projectFiles = ref([] as IProjectFile[]);

const connState = ref(globalThis._ndnd_conn_state);

const busListeners = {
  'project-list': (projs: IProject[]) => (projects.value = projs),
  'project-files': (name: string, files: IProjectFile[]) => {
    activeProjectName.value = name;
    projectFiles.value = files;
  },
  'chat-channels': (chans: IChatChannel[]) => (channels.value = chans),
  'conn-change': () => {
    connState.value = globalThis._ndnd_conn_state;
    if (!connState.value.connected) {
      Toast.info('Disconnected - you are offline');
    }
  },
};

const showNotifBubble = ref(false);

const SIDEBAR_WIDTH_KEY = 'ownly.sidebar.width';
const SIDEBAR_DEFAULT_WIDTH = 230;
const SIDEBAR_MIN_WIDTH = 220;
const SIDEBAR_MAX_WIDTH = 420;

const navRoot = useTemplateRef<HTMLElement>('navRoot');
const sidebarWidth = ref(SIDEBAR_DEFAULT_WIDTH);
const isResizing = ref(false);
const sidebarLeft = ref(0);

const sidebarStyle = computed(() => ({
  width: `${sidebarWidth.value}px`,
  minWidth: `${sidebarWidth.value}px`,
  flex: `0 0 ${sidebarWidth.value}px`,
}));

let interval: ReturnType<typeof setInterval> ;

onMounted(async () => {
  const savedWidth = Number(globalThis.localStorage?.getItem(SIDEBAR_WIDTH_KEY));
  if (Number.isFinite(savedWidth)) {
    sidebarWidth.value = Math.max(SIDEBAR_MIN_WIDTH, Math.min(SIDEBAR_MAX_WIDTH, savedWidth));
  }

  GlobalBus.addListener('project-list', busListeners['project-list']);
  GlobalBus.addListener('project-files', busListeners['project-files']);
  GlobalBus.addListener('chat-channels', busListeners['chat-channels']);
  GlobalBus.addListener('conn-change', busListeners['conn-change']);
  interval = setInterval(() => {
    setNotification();
  },
  250)
});

onUnmounted(() => {
  stopSidebarResize();

  GlobalBus.removeListener('project-list', busListeners['project-list']);
  GlobalBus.removeListener('project-files', busListeners['project-files']);
  GlobalBus.removeListener('chat-channels', busListeners['chat-channels']);
  GlobalBus.removeListener('conn-change', busListeners['conn-change']);
  clearInterval(interval);
});

function startSidebarResize(event: PointerEvent) {
  sidebarLeft.value = navRoot.value?.getBoundingClientRect().left ?? 0;
  isResizing.value = true;
  document.body.style.userSelect = 'none';
  document.body.style.cursor = 'col-resize';

  onSidebarResize(event);
  window.addEventListener('pointermove', onSidebarResize);
  window.addEventListener('pointerup', stopSidebarResize, { once: true });
}

function onSidebarResize(event: PointerEvent) {
  if (!isResizing.value) return;

  const next = Math.max(
    SIDEBAR_MIN_WIDTH,
    Math.min(SIDEBAR_MAX_WIDTH, Math.round(event.clientX - sidebarLeft.value)),
  );
  sidebarWidth.value = next;
}

function stopSidebarResize() {
  if (!isResizing.value) {
    window.removeEventListener('pointermove', onSidebarResize);
    return;
  }

  isResizing.value = false;
  document.body.style.userSelect = '';
  document.body.style.cursor = '';
  window.removeEventListener('pointermove', onSidebarResize);
  globalThis.localStorage?.setItem(SIDEBAR_WIDTH_KEY, String(sidebarWidth.value));
}

function buildVersion() {
  return __BUILD_VERSION__;
}

/** Link to project home page */
function linkProject(project: IProject) {
  return {
    name: 'project',
    params: {
      space: route.params.space,
      project: project.name,
    },
  };
}

/** Link to discussion channel */
function linkDiscuss(channel: IChatChannel) {
  return {
    name: 'discuss',
    params: {
      space: route.params.space,
      channel: channel.name,
    },
  };
}


function setNotification() {
  let wkspName = "/" + route.params.space as string
  while (wkspName.replace("-","/") != wkspName) { // convert dashes to slashes
    wkspName = wkspName.replace("-","/")
  }
  if (_access_requests.filter(a => a[0] == wkspName && a[2] == false).length > 0)
    showNotifBubble.value = true;
  else
    showNotifBubble.value = false;
}
</script>

<style scoped lang="scss">
@use '@/assets/navbar-item.scss';

.main-nav {
  width: 230px;
  min-width: 230px;
  height: 100dvh;
  overflow-y: hidden;
  position: relative;

  display: flex;
  flex-direction: column;
  .top-sheet {
    padding: 10px;
    flex: 1;
    overflow-y: auto;
  }
  .bottom-sheet {
    padding: 8px 10px;
    font-size: 0.9rem;

    a {
      color: #ccc;
    }

    .id-share {
      padding: 2px 12px;
    }

    .connection {
      border-radius: 50px;
      padding: 5px 12px;
      margin-top: 2px;
      color: #ddd;
      cursor: pointer;

      transition: background-color 0.2s ease;
      background-color: rgba(255, 255, 255, 0.05);
      &:hover {
        background-color: rgba(255, 255, 255, 0.1);
      }
    }
  }

  .logo {
    display: block;
    height: 35px;
    margin: 5px;
    margin-bottom: 15px;
  }

  .menu-label {
    color: #bbb;
  }

  :deep(li > a) {
    background-color: transparent;
    color: white;
    display: flex;
    align-items: center;
    //justify-content: space-between;

    &:hover {
      background-color: rgba(255, 255, 255, 0.1);
    }

    &.is-active,
    &.router-link-active {
      background-color: var(--highlight-on-primary-color);
      color: var(--bulma-white-on-scheme);
    }
    .link-inner {
      flex: 1;
      display: flex;
      align-items: center;
    }

    .link-button {
      background: none;
      border: none;
      color: rgba(255, 255, 255, 0.6);
      cursor: pointer;
      width: 20px;
      height: 20px;
      padding: 0;
      border-radius: 4px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      transition: all 0.2s ease;
      opacity: 0;

      &:hover {
      background-color: rgba(255, 255, 255, 0.14);
      color: #fff;
      }
    }

    &:hover .link-button {
      opacity: 1;
    }

  }

  :deep(.project-item > a.router-link-active .link-button),
  :deep(.project-item > a.is-active .link-button) {
    opacity: 1;
  }

  :deep(.project-item > a.project-link) {
    min-height: 30px;
    border-radius: 6px;
    border: 0;
    background-color: transparent;
  }

  :deep(.project-item > a.project-link:hover) {
    background-color: rgba(255, 255, 255, 0.1);
  }

  :deep(.project-item > a.project-link .project-link-inner) {
    gap: 7px;
  }

  :deep(.project-item > a.project-link .project-icon-shell) {
    width: 16px;
    height: 16px;
    border-radius: 4px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto;
  }

  :deep(.project-item > a.project-link .project-name) {
    font-weight: 600;
    letter-spacing: 0.01em;
  }

  :deep(.project-item > .outermost) {
    margin-top: 3px;
    margin-bottom: 8px;
  }

  .project-item + .project-item {
    position: relative;
    margin-top: 6px;
    padding-top: 6px;
  }

  .project-item + .project-item::before {
    content: '';
    position: absolute;
    left: 10px;
    right: 10px;
    top: 0;
    height: 1px;
    background: rgba(255, 255, 255, 0.08);
  }

  :deep(.project-tree-collapse-enter-active),
  :deep(.project-tree-collapse-leave-active) {
    overflow: hidden;
    transition: max-height 0.22s ease, opacity 0.18s ease;
  }

  :deep(.project-tree-collapse-enter-from),
  :deep(.project-tree-collapse-leave-to) {
    max-height: 0;
    opacity: 0;
  }

  :deep(.project-tree-collapse-enter-to),
  :deep(.project-tree-collapse-leave-from) {
    max-height: 1200px;
    opacity: 1;
  }

  @media (prefers-reduced-motion: reduce) {
    :deep(.project-tree-collapse-enter-active),
    :deep(.project-tree-collapse-leave-active) {
      transition: none;
    }
  }

  .sidebar-resizer {
    position: absolute;
    top: 0;
    right: -3px;
    width: 6px;
    height: 100%;
    cursor: col-resize;
    z-index: 30;
    touch-action: none;

    &::before {
      content: '';
      position: absolute;
      top: 0;
      bottom: 0;
      left: 50%;
      width: 1px;
      transform: translateX(-50%);
      background: rgba(255, 255, 255, 0);
      transition: background-color 0.15s ease;
    }
  }

  &:hover .sidebar-resizer::before,
  &.resizing .sidebar-resizer::before {
    background: rgba(255, 255, 255, 0.22);
  }
}
</style>

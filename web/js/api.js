window.API = {
  baseURL: "",
  ws: null,
  wsListeners: [],

  // CSRF и ошибки
  _getCSRF(nodes) {
    const node = nodes.find((n) => n.attributes?.name === "csrf_token");
    return node ? node.attributes.value : "";
  },

  _extractError(data) {
    if (data.ui?.messages?.length > 0) {
      return data.ui.messages[0].text;
    }
    const nodeWithError = data.ui?.nodes?.find((n) => n.messages?.length > 0);
    if (nodeWithError) {
      return nodeWithError.messages[0].text;
    }
    return "Unknown error occurred";
  },

  // Авторизация
  async login(email, password) {
    const flowRes = await fetch("/auth/self-service/login/browser", {
      headers: { Accept: "application/json" },
      credentials: "include",
    });
    if (!flowRes.ok) throw new Error("Failed to initialize login flow");

    const flowData = await flowRes.json();
    const csrfToken = this._getCSRF(flowData.ui.nodes);

    const loginRes = await fetch(
      `/auth/self-service/login?flow=${flowData.id}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          method: "password",
          identifier: email,
          password: password,
          csrf_token: csrfToken,
        }),
      },
    );

    const loginData = await loginRes.json();
    if (!loginRes.ok) throw new Error(this._extractError(loginData));
    return loginData.session;
  },

  async register(email, password, name) {
    const flowRes = await fetch("/auth/self-service/registration/browser", {
      headers: { Accept: "application/json" },
      credentials: "include",
    });
    if (!flowRes.ok) throw new Error("Failed to initialize registration flow");

    const flowData = await flowRes.json();
    const csrfToken = this._getCSRF(flowData.ui.nodes);

    const regRes = await fetch(
      `/auth/self-service/registration?flow=${flowData.id}`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          method: "password",
          password: password,
          traits: { email: email, name: name },
          csrf_token: csrfToken,
        }),
      },
    );

    const regData = await regRes.json();
    if (!regRes.ok) throw new Error(this._extractError(regData));
    return regData;
  },

  async logout() {
    try {
      const flowRes = await fetch("/auth/self-service/logout/browser", {
        headers: { Accept: "application/json" },
        credentials: "include",
      });
      if (!flowRes.ok) throw new Error("Failed to initialize logout flow");

      const flowData = await flowRes.json();
      await fetch(flowData.logout_url, {
        method: "GET",
        headers: { Accept: "application/json" },
        credentials: "include",
      });
    } catch (error) {
      console.error("Logout error:", error);
    } finally {
      this.disconnectWebSocket();
      window.location.href = "/login.html";
    }
  },

  async getCurrentUser() {
    const res = await fetch("/auth/sessions/whoami", {
      headers: { Accept: "application/json" },
      credentials: "include",
    });
    if (!res.ok) throw new Error("Not authenticated");
    return res.json();
  },

  // Планы серверов
  async getPlans(page = 1, pageSize = 10) {
    const res = await fetch(
      `/api/hosting/plans?page=${page}&pageSize=${pageSize}`,
      {
        headers: { Accept: "application/hal+json" },
        credentials: "include",
      },
    );
    if (!res.ok) throw new Error("Failed to fetch plans");
    return res.json();
  },

  async getPlan(planId) {
    const res = await fetch(`/api/hosting/plans/${planId}`, {
      headers: { Accept: "application/hal+json" },
      credentials: "include",
    });
    if (!res.ok) throw new Error("Failed to fetch plan");
    return res.json();
  },

  async createPlan(planData) {
    const res = await fetch("/api/hosting/plans", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      credentials: "include",
      body: JSON.stringify(planData),
    });
    if (!res.ok) {
      const error = await res.json();
      throw new Error(error.message || "Failed to create plan");
    }
    return res.json();
  },

  // Серверы
  async getServers(page = 1, pageSize = 20) {
    const res = await fetch(
      `/api/hosting/servers?page=${page}&pageSize=${pageSize}`,
      {
        headers: { Accept: "application/hal+json" },
        credentials: "include",
      },
    );
    if (!res.ok) throw new Error("Failed to fetch servers");
    return res.json();
  },

  async getServer(serverId) {
    const res = await fetch(`/api/hosting/servers/${serverId}`, {
      headers: { Accept: "application/hal+json" },
      credentials: "include",
    });
    if (!res.ok) throw new Error("Failed to fetch server");
    return res.json();
  },

  async orderServer(planId, name) {
    const res = await fetch("/api/hosting/servers", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/hal+json",
      },
      credentials: "include",
      body: JSON.stringify({ planId, name }),
    });
    if (!res.ok) {
      const error = await res.json();
      throw new Error(error.message || "Failed to order server");
    }
    return res.json();
  },

  async performServerAction(serverId, action) {
    const res = await fetch(`/api/hosting/servers/${serverId}/actions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      credentials: "include",
      body: JSON.stringify({ action }),
    });
    if (!res.ok) {
      const error = await res.json();
      throw new Error(error.message || "Failed to perform action");
    }
    return res.json();
  },

  // WebSocket для уведомлений
  connectWebSocket() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      return;
    }

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/api/notification/ws`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log("WebSocket connected");
    };

    this.ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        this.wsListeners.forEach((listener) => listener(message));
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    this.ws.onclose = () => {
      console.log("WebSocket disconnected");
      // Переподключение через 5 секунд
      setTimeout(() => this.connectWebSocket(), 5000);
    };
  },

  disconnectWebSocket() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  },

  addWebSocketListener(callback) {
    this.wsListeners.push(callback);
  },

  removeWebSocketListener(callback) {
    this.wsListeners = this.wsListeners.filter((l) => l !== callback);
  },
};

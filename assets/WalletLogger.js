/*:
 * @plugindesc Smart Event Sender - Envoie des événements télémétriques à une API Go locale.
 * @help
 * Ce plugin (anciennement WalletLogger) écoute le moteur du jeu et envoie
 * les événements suivants via des requêtes PUT :
 * - shop_open
 * - shop_close
 * - money_add
 * - money_remove
 * - item_buy
 *
 * Les requêtes sont envoyées au serveur local avec un JSON structuré.
 *
 * @param API URL
 * @desc L'URL de l'API locale en Go
 * @default http://localhost:8080/event
 *
 * @param Wallet Variable ID
 * @desc L'ID de la variable gérant l'argent du joueur
 * @default 1
 */

(function () {
  "use strict";

  var parameters = PluginManager.parameters("WalletLogger");
  var apiUrl = parameters["API URL"] || "http://localhost:8080/event";
  var walletVariableId = Number(parameters["Wallet Variable ID"] || 1);

  // ======================================================================
  // Configuration Modulaire
  // ======================================================================
  // Vous pouvez facilement ajouter de nouveaux Common Events ici.
  var CommonEventTriggers = {
    1: "shop_open",
    2: "shop_close",
  };

  // IDs des Common Events correspondants à un achat de boutique
  var ShopPurchaseEvents = [7, 8, 9, 10, 11, 12, 13];

  // ======================================================================
  // Hook : Démarrage du plugin et de la partie
  // ======================================================================
  // Quand le script est lu (le moteur démarre)
  sendTelemetry("plugin_loaded", 0, { status: "ready" });

  // Remplacer DataManager par Scene_Title pour "New Game"
  var _Scene_Title_commandNewGame = Scene_Title.prototype.commandNewGame;
  Scene_Title.prototype.commandNewGame = function () {
    _Scene_Title_commandNewGame.call(this);
    var startBalance = $gameVariables.value(walletVariableId) || 0;
    sendTelemetry("game_started", startBalance, { type: "new_game" });
  };

  // Remplacer DataManager par Scene_Load pour "Load Game"
  var _Scene_Load_onLoadSuccess = Scene_Load.prototype.onLoadSuccess;
  Scene_Load.prototype.onLoadSuccess = function () {
    _Scene_Load_onLoadSuccess.call(this);
    var loadedBalance = $gameVariables.value(walletVariableId) || 0;
    sendTelemetry("game_started", loadedBalance, { type: "load_game" });
  };

  // ======================================================================
  // Fonction d'envoi vers l'API Go
  // ======================================================================
  function sendTelemetry(eventName, balance, extraData) {
    var payload = {
      event: eventName,
      balance: balance,
      details: extraData || {},
    };

    // Mode debug en console
    console.log("[SmartEventSender] Envoi de l'événement:", eventName, payload);

    // Requête asynchrone non-bloquante pour le jeu
    fetch(apiUrl, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    }).catch(function (error) {
      // Silencieux si le serveur n'est pas lancé
      console.warn(
        "[SmartEventSender] Serveur Go non joignable (attendu si éteint) :",
        error.message,
      );
    });
  }

  // ======================================================================
  // Hook 1 & 2 : Interception du démarrage des événements communs
  // ======================================================================
  function handleCommonEventStart(eventId) {
    var currentBalance = $gameVariables.value(walletVariableId);

    if (CommonEventTriggers[eventId]) {
      sendTelemetry(CommonEventTriggers[eventId], currentBalance, {
        commonEventId: eventId,
      });
    }
  }

  var _Game_Interpreter_setupReservedCommonEvent =
    Game_Interpreter.prototype.setupReservedCommonEvent;
  Game_Interpreter.prototype.setupReservedCommonEvent = function () {
    var reservedEventId = $gameTemp.isCommonEventReserved()
      ? $gameTemp._commonEventId
      : 0;
    var result = _Game_Interpreter_setupReservedCommonEvent.call(this);

    if (result && reservedEventId > 0) {
      this._wl_commonEventId = reservedEventId;
      handleCommonEventStart(reservedEventId);
    }
    return result;
  };

  var _Game_Interpreter_command117 = Game_Interpreter.prototype.command117;
  Game_Interpreter.prototype.command117 = function () {
    var calledEventId = this._params[0];
    var result = _Game_Interpreter_command117.call(this);
    if (this._childInterpreter) {
      this._childInterpreter._wl_commonEventId = calledEventId;
    }
    handleCommonEventStart(calledEventId);
    return result;
  };

  // ======================================================================
  // Hook 3 : Suivi du contexte lors de la commande 122 (Variables)
  // ======================================================================
  var _Game_Interpreter_command122 = Game_Interpreter.prototype.command122;
  Game_Interpreter.prototype.command122 = function () {
    if ($gameTemp) {
      $gameTemp._wl_activeInterpreter = this;
    }
    var result = _Game_Interpreter_command122.call(this);
    if ($gameTemp) {
      $gameTemp._wl_activeInterpreter = null;
    }
    return result;
  };

  // ======================================================================
  // Hook 4 : Interception du changement de portefeuille
  // ======================================================================
  var _Game_Variables_setValue = Game_Variables.prototype.setValue;
  Game_Variables.prototype.setValue = function (variableId, value) {
    var oldValue = this.value(variableId);
    _Game_Variables_setValue.call(this, variableId, value);

    if (variableId === walletVariableId) {
      var diff = this.value(variableId) - oldValue;
      if (diff === 0) return; // Pas de changement réel

      var currentBalance = this.value(variableId);

      // Récupérer le contexte (Dans quel Common Event sommes-nous ?)
      var interpreter =
        $gameTemp && $gameTemp._wl_activeInterpreter
          ? $gameTemp._wl_activeInterpreter
          : null;
      var currentCeId = interpreter ? interpreter._wl_commonEventId || 0 : 0;

      var details = {
        diff: diff,
        commonEventId: currentCeId,
      };

      // Classification intelligente de l'événement
      if (diff < 0 && ShopPurchaseEvents.includes(currentCeId)) {
        sendTelemetry("item_buy", currentBalance, details);
      } else if (diff < 0) {
        sendTelemetry("money_remove", currentBalance, details);
      } else if (diff > 0) {
        sendTelemetry("money_add", currentBalance, details);
      }
    }
  };
})();

/*
  To add in plugin.js
  {
    name: "WalletLogger",
    status: true,
    description: "Logs wallet changes and shop transactions to the console.",
    parameters: {
      "Wallet Variable ID": "1",
      "Log Variable Changes": "true",
      "Log Shop Actions": "true",
    },
  }
*/

/*:
 * @plugindesc Smart Event Sender - Sends telemetry events to a local Go API.
 * @help
 * This plugin (formerly WalletLogger) listens to the game engine and sends
 * the following events via PUT requests:
 * - shop_open
 * - shop_close
 * - money_add
 * - money_remove
 * - item_buy
 *
 * Requests are sent to the local server with structured JSON.
 *
 * @param API URL
 * @desc The URL of the local Go API
 * @default http://localhost:69420/event
 *
 * @param Wallet Variable ID
 * @desc The ID of the variable managing the player's money
 * @default 1
 */

(function () {
  "use strict";

  var parameters = PluginManager.parameters("WalletLogger");
  var apiUrl = parameters["API URL"] || "http://localhost:69420/event";
  var walletVariableId = Number(parameters["Wallet Variable ID"] || 1);

  // ======================================================================
  // Modular Configuration
  // ======================================================================
  // You can easily add new Common Events here.
  var CommonEventTriggers = {
    1: "shop_open",
    2: "shop_close",
  };

  // Common Event IDs corresponding to a shop purchase
  var ShopPurchaseEvents = [7, 8, 9, 10, 11, 12, 13];

  // ======================================================================
  // Hook: Plugin and game startup
  // ======================================================================
  // When the script is read (the engine starts)
  sendTelemetry("plugin_loaded", 0, { status: "ready" });

  // Replace DataManager with Scene_Title for "New Game"
  var _Scene_Title_commandNewGame = Scene_Title.prototype.commandNewGame;
  Scene_Title.prototype.commandNewGame = function () {
    _Scene_Title_commandNewGame.call(this);
    var startBalance = $gameVariables.value(walletVariableId) || 0;
    sendTelemetry("game_started", startBalance, { type: "new_game" });
  };

  // Replace DataManager with Scene_Load for "Load Game"
  var _Scene_Load_onLoadSuccess = Scene_Load.prototype.onLoadSuccess;
  Scene_Load.prototype.onLoadSuccess = function () {
    _Scene_Load_onLoadSuccess.call(this);
    var loadedBalance = $gameVariables.value(walletVariableId) || 0;
    sendTelemetry("game_started", loadedBalance, { type: "load_game" });
  };

  // ======================================================================
  // Send function to the Go API
  // ======================================================================
  function sendTelemetry(eventName, balance, extraData) {
    var payload = {
      event: eventName,
      balance: balance,
      details: extraData || {},
    };

    // Console debug mode
    console.log("[SmartEventSender] Sending event:", eventName, payload);

    // Non-blocking async request for the game
    fetch(apiUrl, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    }).catch(function (error) {
      // Silent if the server is not running
      console.warn(
        "[SmartEventSender] Go server unreachable (expected if off):",
        error.message,
      );
    });
  }

  // ======================================================================
  // Hook 1 & 2: Intercept common event start
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
  // Hook 3: Track context during command 122 (Variables)
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
  // Hook 4: Intercept wallet changes
  // ======================================================================
  var _Game_Variables_setValue = Game_Variables.prototype.setValue;
  Game_Variables.prototype.setValue = function (variableId, value) {
    var oldValue = this.value(variableId);
    _Game_Variables_setValue.call(this, variableId, value);

    if (variableId === walletVariableId) {
      var diff = this.value(variableId) - oldValue;
      if (diff === 0) return; // No real change

      var currentBalance = this.value(variableId);

      // Get context (which Common Event are we in?)
      var interpreter =
        $gameTemp && $gameTemp._wl_activeInterpreter
          ? $gameTemp._wl_activeInterpreter
          : null;
      var currentCeId = interpreter ? interpreter._wl_commonEventId || 0 : 0;

      var details = {
        diff: diff,
        commonEventId: currentCeId,
      };

      // Smart event classification
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

#include "cashpointedit.h"

#include <QtCore/QDebug>
#include <QtCore/QJsonObject>
#include <QtCore/QJsonParseError>

#include "../serverapi.h"

CashPointEdit::CashPointEdit(CashPointSqlModel *model, QJSValue callback)
    : CashPointRequest(model, callback)
{
    registerStepHandlers({
                             STEP_HANDLER(editCashpoint)
                         });
}

bool CashPointEdit::fromJson(const QJsonObject &json)
{
    const QJsonValue id = json["id"];

    const QJsonValue type = json["type"];
    const QJsonValue bankId = json["bank_id"];
    const QJsonValue townId = json["town_id"];
    const QJsonValue longitude = json["longitude"];
    const QJsonValue latitude = json["latitide"];
    //const QJsonValue address = json["address"];
    const QJsonValue addressComment = json["address_comment"];
    //const QJsonValue metroName = json["metro_name"];
    const QJsonValue freeAccess = json["free_access"];
    const QJsonValue mainOffice = json["main_office"];
    const QJsonValue withoutWeekend = json["without_weekend"];
    const QJsonValue roundTheClock = json["round_the_clock"];
    const QJsonValue worksAsShop = json["works_as_shop"];
    const QJsonValue schedule = json["schedule"];
    const QJsonValue tel = json["tel"];
    const QJsonValue additional = json["additional"];
    const QJsonValue currency = json["currency"];
    const QJsonValue cashIn = json["cash_in"];

    CHECK_JSON_TYPE_NUMBER_STRICT(id)

    CHECK_JSON_TYPE_STRING(type)
    CHECK_JSON_TYPE_NUMBER(bankId)
    CHECK_JSON_TYPE_NUMBER(townId)
    CHECK_JSON_TYPE_NUMBER(longitude)
    CHECK_JSON_TYPE_NUMBER(latitude)
    //CHECK_JSON_TYPE_STRING(address)
    CHECK_JSON_TYPE_STRING(addressComment)
    //CHECK_JSON_TYPE_STRING_STRICT(metroName)
    CHECK_JSON_TYPE_BOOL(freeAccess)
    CHECK_JSON_TYPE_BOOL(mainOffice)
    CHECK_JSON_TYPE_BOOL(withoutWeekend)
    CHECK_JSON_TYPE_BOOL(roundTheClock)
    CHECK_JSON_TYPE_BOOL(worksAsShop)
    CHECK_JSON_TYPE_OBJECT(schedule)
    CHECK_JSON_TYPE_STRING(tel)
    CHECK_JSON_TYPE_STRING(additional)
    CHECK_JSON_TYPE_ARRAY_STRICT(currency)
    CHECK_JSON_TYPE_BOOL(cashIn)

    data = json;
    return true;
}

void CashPointEdit::editCashpoint(ServerApiPtr api, quint32 leftAttempts)
{
    const int step = 0;

    QJsonObject reqData;
    reqData["user_id"] = 0;
    reqData["data"] = data;
    //int id = data["id"].toDouble();

    mId = api->sendRequest(//"/cashpoint/" + QString::number(id), data,
                           "/cashpoint", reqData,
    [this, api, leftAttempts](ServerApi::RequestStatusCode reqCode, ServerApi::HttpStatusCode httpCode, const QByteArray &data) {
        if (!isRunning()) {
            if (isDisposing()) {
                deleteLater();
            }
            return;
        }

        if (reqCode == ServerApi::RSC_Timeout) {
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        const QString errText = trUtf8("Cannot edit existing cashpoint");
        if (reqCode != ServerApi::RSC_Ok) {
            emitError(ServerApi::requestStatusCodeText(reqCode));
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitError(trUtf8("Server http response: %1").arg(httpCode));
            emitStepFinished(*api, step, false, errText);
            return;
        }

        /*QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitError("CashPointEitd: server response json parse error: " + err.errorString());
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (!json.isObject()) {
            emitError("CashPointEdit: response is not an object!");
            emitStepFinished(*api, step, false, errText);
            return;
        }

        const QJsonObject obj = json.object();
        if (!obj["cash_points"].isArray()) {
            emitError("CashPointEdit: cash_points field is not an array!");
            emitStepFinished(*api, step, false, errText);
            return;
        }*/

        mResponse = new CashPointResponse;
        mResponse->type = CashPointResponse::EditResult;
        mResponse->message = trUtf8("Cashpoint successfully edited");
        emitStepFinished(*api, step, true, "");
        emitResponseReady(true);
    });
}

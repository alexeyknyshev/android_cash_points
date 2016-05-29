#include "cashpointcreate.h"

#include <QtCore/QJsonObject>
#include <QtCore/QJsonArray>
#include <QtCore/QJsonParseError>
#include <QtCore/QDebug>

#include "../serverapi.h"

CashPointCreate::CashPointCreate(CashPointSqlModel *model, QJSValue callback)
    : CashPointRequest(model, callback)
{
    registerStepHandlers({
                             STEP_HANDLER(createCashpoint)
                         });
}

bool CashPointCreate::fromJson(const QJsonObject &json)
{
    const QJsonValue type = json["type"];
    const QJsonValue bankId = json["bank_id"];
    const QJsonValue townId = json["town_id"];
    const QJsonValue longitude = json["longitude"];
    const QJsonValue latitude = json["latitude"];
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

    CHECK_JSON_TYPE_STRING_STRICT(type)
    CHECK_JSON_TYPE_NUMBER_STRICT(bankId)
    CHECK_JSON_TYPE_NUMBER_STRICT(townId)
    CHECK_JSON_TYPE_NUMBER_STRICT(longitude)
    CHECK_JSON_TYPE_NUMBER_STRICT(latitude)
    //CHECK_JSON_TYPE_STRING_STRICT(address)
    CHECK_JSON_TYPE_STRING_STRICT(addressComment)
    //CHECK_JSON_TYPE_STRING_STRICT(metroName)
    CHECK_JSON_TYPE_BOOL_STRICT(freeAccess)
    CHECK_JSON_TYPE_BOOL_STRICT(mainOffice)
    CHECK_JSON_TYPE_BOOL_STRICT(withoutWeekend)
    CHECK_JSON_TYPE_BOOL_STRICT(roundTheClock)
    CHECK_JSON_TYPE_BOOL_STRICT(worksAsShop)
    CHECK_JSON_TYPE_OBJECT_STRICT(schedule)
    CHECK_JSON_TYPE_STRING_STRICT(tel)
    CHECK_JSON_TYPE_STRING_STRICT(additional)
    CHECK_JSON_TYPE_ARRAY_STRICT(currency)
    CHECK_JSON_TYPE_BOOL_STRICT(cashIn)

    data = json;
    return true;
}

void CashPointCreate::createCashpoint(ServerApiPtr api, quint32 leftAttempts)
{
    const int step = 0;

    QJsonObject reqData;
    reqData["user_id"] = 0;
    reqData["data"] = data;

    mId = api->sendRequest("/cashpoint", reqData,
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

        const QString errText = trUtf8("Cannot add new cashpoint to system");
        if (reqCode != ServerApi::RSC_Ok) {
            emitError(ServerApi::requestStatusCodeText(reqCode));
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (httpCode != ServerApi::HSC_Ok) {
            qWarning() << "Server request http code: " << httpCode;
            emitUpdate(leftAttempts - 1, step);
            return;
        }

        bool ok = false;
        data.toLongLong(&ok);

        QString text;
        if (ok) {
            text = trUtf8("Cashpoint successfully added");
        } else {
            text = trUtf8("Cashpoint adding failed!");
        }
        emitStepFinished(*api, step, ok, text);

        mResponse = new CashPointResponse;
        mResponse->type = CashPointResponse::CreateResult;
        emitResponseReady(true);
/*
        QJsonParseError err;
        const QJsonDocument json = QJsonDocument::fromJson(data, &err);
        if (err.error != QJsonParseError::NoError) {
            emitError("CashPointCreate: server response json parse error: " + err.errorString());
            emitStepFinished(*api, step, false, errText);
            return;
        }

        if (!json.isObject()) {
            emitError("CashPointCreate: response is not an object!");
            emitStepFinished(*api, step, false, errText);
            return;
        }

        const QJsonObject obj = json.object();
        if (!obj["cash_points"].isArray()) {
            emitError("CashPointCreate: cash_points field is not an array!");
            emitStepFinished(*api, step, false, errText);
            return;
        }

        mResponse = new CashPointResponse;
        mResponse->type = CashPointResponse::CreateResult;
        bool ok = false;

        const QJsonArray arr = obj["cash_points"].toArray();
        const auto end = arr.constEnd();
        for (auto it = arr.constBegin(); it != end; it++) {
            const QJsonValue &val = *it;
            const int id = val.toInt();
            if (id > 0) {
                ok = true;
                mResponse->addCashPointData(this->data);
            }
        }

        if (ok) {
            mResponse->message = trUtf8("Cashpoint successfully added");
        } else {
            mResponse->message = trUtf8("Cashpoint adding failed!");
        }
        emitResponseReady(true);*/
    });
}

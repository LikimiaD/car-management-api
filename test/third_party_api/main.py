from typing import List

from fastapi import FastAPI, HTTPException
from models import Car, CarDatabase


app = FastAPI(title="Car Info API", version="0.0.1")

car_db = CarDatabase()


@app.get("/api/info", response_model=Car, responses={404: {"description": "Car not found"}})
async def get_car_info(regNum: str):
    car = car_db.get_car_info(regNum)
    if car is None:
        raise HTTPException(status_code=404, detail="Car not found")
    return car


@app.get("/api/all_reg_numbers", response_model=List[str])
async def get_all_reg_numbers():
    return car_db.get_all_reg_numbers()
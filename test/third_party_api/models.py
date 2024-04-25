from typing import Optional
from faker import Faker
from pydantic import BaseModel
from random import choice

fake = Faker()


class People(BaseModel):
    name: str
    surname: str
    patronymic: Optional[str] = None


class Car(BaseModel):
    regNum: str
    mark: str
    model: str
    year: int
    owner: People


class CarDatabase:
    def __init__(self):
        self.cars = {}
        self.populate_cars(100)

    def populate_cars(self, count):
        marks_models = [
            ("Toyota", "Corolla"),
            ("Ford", "Focus"),
            ("Nissan", "Altima"),
            ("Chevrolet", "Impala"),
            ("Honda", "Civic")
        ]
        for _ in range(count):
            reg_num = fake.bothify(text='??###??###')
            mark, model = choice(marks_models)
            car = Car(
                regNum=reg_num,
                mark=mark,
                model=model,
                year=fake.year(),
                owner=People(
                    name=fake.first_name(),
                    surname=fake.last_name(),
                    patronymic=fake.first_name()
                )
            )
            self.cars[reg_num] = car

    def get_car_info(self, reg_num: str):
        if reg_num in self.cars:
            return self.cars[reg_num]
        else:
            return None

    def get_all_reg_numbers(self):
        return list(self.cars.keys())
